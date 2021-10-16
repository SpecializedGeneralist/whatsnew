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

// Client acts as a client to the HNSW service and contains essential
// functionalities for managing indices, and inserting and searching vectors.
//
// This client is designed to handle vectors related to WebArticle models.
//
// Inserting Vectors
//
// New vectors can be inserted calling Client.Insert.
//
// The given ID is explicitly assigned to the stored vector (the HNSW service's
// auto-ID feature is always disabled).
//
// Each vector must be associated to a specific datetime which, in
// the default implementation, corresponds to WebArticle.PublishDate.
//
// At insertion time, this datetime is converted to UTC and truncated (rounded
// down) to the UTC day. The resulting formatted date "YYYY-MM-DD" is appended
// to the configuration's NamePrefix and used as destination index name.
//
// If an index with this final name does not exist, a new one is created
// using the given configuration settings.
//
// The index is always flushed after each insertion.
//
// Searching for Similar Vectors
//
// Given a reference vector, the function Client.SearchKNN allows searching
// for similar or duplicate vectors across one or more indices.
//
// SearchParams.From and To datetimes are converted to UTC and truncated
// to the UTC day (just as already described above). The resulting range of
// UTC days, leading and trailing dates included, is used to generate the set
// of names of indices among which similar vectors must be searched (once again,
// each day-date is formatted as "YYYY-MM-DD" and appended to the prefix).
//
// If an index in that range does not exist, it is considered as an empty index,
// just not producing any matching result (hit).
//
// The raw hits from each existing index (if any) are filtered, keeping only
// the ones whose Distance from the initial vector is less than or equal to
// the specified SearchParams.DistanceThreshold.
//
// The full set of filtered results is finally converted to a sorted list of
// hits and returned.
//
// Indices Cache
//
// Existing index names are cached locally to reduce the number of requests.
// If you modify the remote HNSW indices, manually or with other tools, be sure
// to restart the processes which are using this Client to prevent errors.
type Client struct {
	cli          grpcapi.ServerClient
	conf         config.HNSWIndex
	indicesCache sets.StringSet
}

// Hit is a single search result.
type Hit struct {
	ID       uint
	Distance float32
}

// LessThan reports whether h has a smaller Distance value than other.
// If the two distances are identical, the ID values are compared instead,
// in order to preserve stability for sorting operations (see Hits).
func (h Hit) LessThan(other Hit) bool {
	if h.Distance == other.Distance {
		return h.ID < other.ID
	}
	return h.Distance < other.Distance
}

// Hits is a sortable list of Hit.
type Hits []Hit

// Len is the number of elements in the collection.
func (h Hits) Len() int {
	return len(h)
}

// Less reports whether the element with index i must sort before the element
// with index j.
func (h Hits) Less(i, j int) bool {
	return h[i].LessThan(h[j])
}

// Swap swaps the elements with indexes i and j.
func (h Hits) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// SearchParams provides parameters for Client.SearchKNN.
type SearchParams struct {
	From              time.Time
	To                time.Time
	Vector            []float32
	DistanceThreshold float32
}

const indexNameTimeLayout = "2006-01-02"
const day = 24 * time.Hour

// New creates a new Client.
func New(conn *grpc.ClientConn, conf config.HNSWIndex) *Client {
	return &Client{
		cli:          grpcapi.NewServerClient(conn),
		conf:         conf,
		indicesCache: sets.NewStringSet(),
	}
}

// Insert inserts a new vector.
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

// SearchKNN performs k-nearest-neighbors search over one or more indices,
// filtering, merging and sorting all results.
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

// IndicesOlderThan returns a list of names of indices whose WebArticle's
// publishing date is older than the given Time.
func (c *Client) IndicesOlderThan(ctx context.Context, t time.Time) ([]string, error) {
	err := c.fetchIndices(ctx)
	if err != nil {
		return nil, err
	}

	upperIndexName := c.dailyIndexName(t)
	expectedLen := len(c.conf.NamePrefix) + len(indexNameTimeLayout)

	indices := make([]string, 0, len(c.indicesCache))
	for index := range c.indicesCache {
		// Because of how the dates are formatted, we can simply compare
		// the strings to get the older indices, without involving
		// time parsing.
		if len(index) == expectedLen &&
			strings.HasPrefix(index, c.conf.NamePrefix) &&
			index < upperIndexName {
			indices = append(indices, index)
		}
	}

	return indices, nil
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
			ID:       id,
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
