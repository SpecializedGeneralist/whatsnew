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
	"strings"
	"time"
)

type Client struct {
	cli          grpcapi.ServerClient
	conf         config.HNSWIndex
	indicesCache sets.StringSet
}

func New(conn *grpc.ClientConn, conf config.HNSWIndex) *Client {
	return &Client{
		cli:          grpcapi.NewServerClient(conn),
		conf:         conf,
		indicesCache: sets.NewStringSet(),
	}
}

func (c *Client) Insert(ctx context.Context, id uint, t time.Time, vec []float32) error {
	indexName := c.indexName(t)

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

const indexNameTimeLayout = "2006-01-02"

func (c *Client) indexName(t time.Time) string {
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
