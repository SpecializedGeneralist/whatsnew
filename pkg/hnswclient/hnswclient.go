// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hnswclient

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/sets"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	cli          grpcapi.ServerClient
	conf         config.HNSWIndex
	indicesCache sets.StringSet
}

type Hit struct {
	ID       uint
	Distance float32
}

func (h Hit) LessThan(other Hit) bool {
	if h.Distance == other.Distance {
		return h.ID < other.ID
	}
	return h.Distance < other.Distance
}

type Hits []Hit

func (h Hits) Len() int           { return len(h) }
func (h Hits) Less(i, j int) bool { return h[i].LessThan(h[j]) }
func (h Hits) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

type SearchParams struct {
	From              time.Time
	To                time.Time
	Vector            []float32
	DistanceThreshold float32
}

const indexNameTimeLayout = "2006-01-02"
const day = 24 * time.Hour

func New(conn *grpc.ClientConn, conf config.HNSWIndex) *Client {
	return &Client{
		cli:          grpcapi.NewServerClient(conn),
		conf:         conf,
		indicesCache: sets.NewStringSet(),
	}
}

func (c *Client) Insert(ctx context.Context, id uint, t time.Time, vec []float32) error {
	indexName := c.dailyIndexName(t)

	err := c.ensureIndexExists(ctx, indexName)
	if err != nil {
		return err
	}

	_, err = c.cli.InsertVectorWithId(ctx, &grpcapi.InsertVectorWithIdRequest{
		IndexName: indexName,
		Id:        int32(id),
		Vector:    &grpcapi.Vector{Value: vec},
	})
	if err != nil {
		return fmt.Errorf("error inserting HNSW vector: %w", err)
	}

	_, err = c.cli.FlushIndex(ctx, &grpcapi.FlushRequest{IndexName: indexName})
	if err != nil {
		return fmt.Errorf("error flushing HNSW index %#v: %w", indexName, err)
	}
	return nil
}

func (c *Client) SearchKNN(ctx context.Context, params SearchParams) (Hits, error) {
	cacheUpdated := false
	hits := make(Hits, 0)

	indexNames := c.dailyIndexNameRange(params.From, params.To)
	for _, indexName := range indexNames {
		if !c.indicesCache.Has(indexName) {
			if cacheUpdated {
				continue
			}
			if err := c.fetchIndices(ctx); err != nil {
				return nil, err
			}
			cacheUpdated = true
			if !c.indicesCache.Has(indexName) {
				continue
			}
		}

		grpcHits, err := c.searchKNN(ctx, indexName, params.Vector)
		if err != nil {
			return nil, err
		}

		newHits, err := filterAndConvertHits(grpcHits, params.DistanceThreshold)
		if err != nil {
			return nil, err
		}

		hits = append(hits, newHits...)
	}

	sort.Sort(hits)
	return hits, nil
}

func (c *Client) dailyIndexNameRange(from, to time.Time) []string {
	first := from.UTC().Truncate(day)
	last := to.UTC().Truncate(day)

	length := int(last.Sub(first)/day) + 1
	names := make([]string, 0, length)

	for t := first; !t.After(last); t = t.Add(day) {
		names = append(names, c.dailyIndexName(t))
	}
	return names
}

func (c *Client) searchKNN(ctx context.Context, indexName string, vec []float32) ([]*grpcapi.Hit, error) {
	reply, err := c.cli.SearchKNN(ctx, &grpcapi.SearchRequest{
		IndexName: indexName,
		Vector:    &grpcapi.Vector{Value: vec},
		K:         c.conf.MaxElements,
	})
	if err != nil {
		return nil, fmt.Errorf("search-KNN error: %w", err)
	}
	return reply.GetHits(), nil
}

func filterAndConvertHits(grpcHits []*grpcapi.Hit, distanceThreshold float32) (Hits, error) {
	hits := make(Hits, 0)
	for _, hit := range grpcHits {
		if hit.Distance > distanceThreshold {
			continue
		}
		id, err := parseID(hit.Id)
		if err != nil {
			return nil, err
		}
		hits = append(hits, Hit{
			ID:       uint(id),
			Distance: hit.Distance,
		})
	}
	return hits, nil
}

func parseID(id string) (uint, error) {
	i, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing ID %#v: %w", id, err)
	}
	return uint(i), nil
}

func (c *Client) dailyIndexName(t time.Time) string {
	suffix := t.UTC().Format(indexNameTimeLayout)
	return fmt.Sprintf("%s%s", c.conf.NamePrefix, suffix)
}

func (c *Client) ensureIndexExists(ctx context.Context, indexName string) error {
	exists, err := c.indexExists(ctx, indexName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return c.createIndex(ctx, indexName)
}

func (c *Client) indexExists(ctx context.Context, indexName string) (bool, error) {
	if c.indicesCache.Has(indexName) {
		return true, nil
	}
	err := c.fetchIndices(ctx)
	if err != nil {
		return false, err
	}
	return c.indicesCache.Has(indexName), nil
}

func (c *Client) fetchIndices(ctx context.Context) error {
	reply, err := c.cli.Indices(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("error getting HNSW indices: %w", err)
	}
	indices := reply.GetIndices()
	for _, indexName := range indices {
		c.indicesCache.Add(indexName)
	}
	return nil
}

func (c *Client) createIndex(ctx context.Context, indexName string) error {
	_, err := c.cli.CreateIndex(ctx, &grpcapi.CreateIndexRequest{
		IndexName:      indexName,
		Dim:            c.conf.Dim,
		EfConstruction: c.conf.EfConstruction,
		M:              c.conf.M,
		MaxElements:    c.conf.MaxElements,
		Seed:           c.conf.Seed,
		SpaceType:      grpcapi.CreateIndexRequest_SpaceType(c.conf.SpaceType),
		AutoId:         false,
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("error creating HNSW index %#v: %w", indexName, err)
	}
	c.indicesCache.Add(indexName)
	return nil
}
