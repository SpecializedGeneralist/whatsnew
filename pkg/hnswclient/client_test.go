// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hnswclient

import (
	"context"
	"errors"
	pb "github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
	"time"
)

func Test_Insert(t *testing.T) {
	t.Parallel()

	tz, llErr := time.LoadLocation("Europe/Berlin")
	require.NoError(t, llErr)
	// This date is "2000-02-01" in UTC
	tm := time.Date(2000, time.February, 2, 0, 0, 0, 0, tz)

	t.Run("successful insertion", func(t *testing.T) {
		t.Parallel()

		indicesCallsCount := 0
		createIndexCallsCount := 0
		insertVectorWithIDCallsCount := 0

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				indicesCallsCount++
				return &pb.IndicesReply{Indices: nil}, nil
			},
			createIndex: func(req *pb.CreateIndexRequest) error {
				createIndexCallsCount++
				assert.Equal(t, "test_2000-02-01", req.GetIndexName())
				assert.EqualValues(t, 3, req.GetDim())
				assert.EqualValues(t, 200, req.GetEfConstruction())
				assert.EqualValues(t, 48, req.GetM())
				assert.EqualValues(t, 100_000, req.GetMaxElements())
				assert.EqualValues(t, 42, req.GetSeed())
				assert.Equal(t, pb.CreateIndexRequest_COSINE, req.GetSpaceType())
				assert.False(t, req.GetAutoId())
				return nil
			},
			insertVectorWithID: func(req *pb.InsertVectorWithIdRequest) (*pb.InsertVectorWithIdReply, error) {
				insertVectorWithIDCallsCount++
				assert.Equal(t, "test_2000-02-01", req.GetIndexName())
				assert.EqualValues(t, 123, req.GetId())
				assert.Equal(t, []float32{1, 2, 3}, req.GetVector().GetValue())
				return &pb.InsertVectorWithIdReply{Took: 1}, nil
			},
		}
		c := New(tc, testingConfig)

		err := c.Insert(context.Background(), 123, tm, []float32{1, 2, 3})
		assert.NoError(t, err)

		assert.Equal(t, 1, indicesCallsCount)
		assert.Equal(t, 1, createIndexCallsCount)
		assert.Equal(t, 1, insertVectorWithIDCallsCount)
	})

	t.Run("Indices request error", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return nil, errTesting
			},
		}
		c := New(tc, testingConfig)

		err := c.Insert(context.Background(), 123, tm, []float32{1, 2, 3})
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
	})

	t.Run("CreateIndex request error", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return &pb.IndicesReply{Indices: nil}, nil
			},
			createIndex: func(req *pb.CreateIndexRequest) error {
				return errTesting
			},
		}
		c := New(tc, testingConfig)

		err := c.Insert(context.Background(), 123, tm, []float32{1, 2, 3})
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
	})

	t.Run("InsertVectorWithId request error", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return &pb.IndicesReply{Indices: nil}, nil
			},
			createIndex: func(req *pb.CreateIndexRequest) error {
				return nil
			},
			insertVectorWithID: func(req *pb.InsertVectorWithIdRequest) (*pb.InsertVectorWithIdReply, error) {
				return nil, errTesting
			},
		}
		c := New(tc, testingConfig)

		err := c.Insert(context.Background(), 123, tm, []float32{1, 2, 3})
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
	})

	t.Run("CreateIndex is not called if the index already exists", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return &pb.IndicesReply{Indices: []string{"test_2000-02-01"}}, nil
			},
			createIndex: func(req *pb.CreateIndexRequest) error {
				assert.FailNow(t, "CreateIndex must not be invoked")
				return nil
			},
			insertVectorWithID: func(req *pb.InsertVectorWithIdRequest) (*pb.InsertVectorWithIdReply, error) {
				return &pb.InsertVectorWithIdReply{Took: 1}, nil
			},
		}
		c := New(tc, testingConfig)

		err := c.Insert(context.Background(), 123, tm, []float32{1, 2, 3})
		assert.NoError(t, err)
	})

	t.Run("indices are cached so that CreateIndex is called only once", func(t *testing.T) {
		t.Parallel()

		indicesCallsCount := 0
		insertVectorWithIDCallsCount := 0

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				indicesCallsCount++
				return &pb.IndicesReply{Indices: []string{"test_2000-02-01"}}, nil
			},
			insertVectorWithID: func(req *pb.InsertVectorWithIdRequest) (*pb.InsertVectorWithIdReply, error) {
				insertVectorWithIDCallsCount++
				return &pb.InsertVectorWithIdReply{Took: 1}, nil
			},
		}
		c := New(tc, testingConfig)

		err := c.Insert(context.Background(), 123, tm, []float32{1, 2, 3})
		assert.NoError(t, err)

		err = c.Insert(context.Background(), 456, tm, []float32{4, 5, 6})
		assert.NoError(t, err)

		assert.Equal(t, 1, indicesCallsCount)
		assert.Equal(t, 2, insertVectorWithIDCallsCount)
	})
}

func TestClient_SearchKNN(t *testing.T) {
	t.Parallel()

	tz, llErr := time.LoadLocation("Europe/Berlin")
	require.NoError(t, llErr)

	t.Run("successful search on one index", func(t *testing.T) {
		t.Parallel()

		indicesCallsCount := 0
		searchKNNCallsCount := 0

		// This date is "2000-02-01" in UTC
		tm := time.Date(2000, time.February, 2, 0, 0, 0, 0, tz)

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				indicesCallsCount++
				return &pb.IndicesReply{Indices: []string{"test_2000-02-01"}}, nil
			},
			searchKNN: func(req *pb.SearchRequest) (*pb.SearchKNNReply, error) {
				searchKNNCallsCount++
				assert.Equal(t, "test_2000-02-01", req.GetIndexName())
				assert.Equal(t, []float32{1, 2, 3}, req.GetVector().GetValue())
				assert.EqualValues(t, 100_000, req.GetK())

				return &pb.SearchKNNReply{
					Hits: []*pb.Hit{
						{Id: "33", Distance: 0.6},
						{Id: "22", Distance: 0.4},
						{Id: "11", Distance: 0.2},
					},
					Took: 2,
				}, nil
			},
		}
		c := New(tc, testingConfig)

		hits, err := c.SearchKNN(context.Background(), SearchParams{
			From:              tm,
			To:                tm,
			Vector:            []float32{1, 2, 3},
			DistanceThreshold: 0.5,
		})
		assert.NoError(t, err)
		assert.Equal(t, Hits{
			{ID: 11, Distance: 0.2},
			{ID: 22, Distance: 0.4},
		}, hits)

		assert.Equal(t, 1, indicesCallsCount)
		assert.Equal(t, 1, searchKNNCallsCount)
	})

	t.Run("successful search on multiple indices", func(t *testing.T) {
		t.Parallel()

		indicesCallsCount := 0
		searchKNNCallsCount := map[int]int{1: 0, 2: 0, 4: 0}

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				indicesCallsCount++
				return &pb.IndicesReply{Indices: []string{
					"test_2000-02-01",
					"test_2000-02-02",
					"test_2000-02-04",
				}}, nil
			},
			searchKNN: func(req *pb.SearchRequest) (*pb.SearchKNNReply, error) {
				assert.Equal(t, []float32{1, 2, 3}, req.GetVector().GetValue())
				assert.EqualValues(t, 100_000, req.GetK())

				switch req.GetIndexName() {
				case "test_2000-02-01":
					searchKNNCallsCount[1]++
					return &pb.SearchKNNReply{
						Hits: []*pb.Hit{
							{Id: "1", Distance: 0.1},
							{Id: "2", Distance: 0.3},
							{Id: "3", Distance: 0.6},
						},
						Took: 2,
					}, nil
				case "test_2000-02-02":
					searchKNNCallsCount[2]++
					return &pb.SearchKNNReply{
						Hits: []*pb.Hit{
							{Id: "4", Distance: 0.2},
							{Id: "5", Distance: 0.4},
							{Id: "6", Distance: 0.7},
						},
						Took: 2,
					}, nil
				case "test_2000-02-04":
					searchKNNCallsCount[4]++
					return &pb.SearchKNNReply{
						Hits: []*pb.Hit{
							{Id: "7", Distance: 0.3},
							{Id: "8", Distance: 0.8},
							{Id: "9", Distance: 0.9},
						},
						Took: 2,
					}, nil
				default:
					assert.FailNow(t, "unexpected SearchKNN call for index %#v", req.GetIndexName())
					return nil, nil
				}
			},
		}
		c := New(tc, testingConfig)

		hits, err := c.SearchKNN(context.Background(), SearchParams{
			// From "2000-02-01" UTC to "2000-02-04" UTC
			From:              time.Date(2000, time.February, 2, 0, 0, 0, 0, tz),
			To:                time.Date(2000, time.February, 5, 0, 0, 0, 0, tz),
			Vector:            []float32{1, 2, 3},
			DistanceThreshold: 0.5,
		})
		assert.NoError(t, err)
		assert.Equal(t, Hits{
			{ID: 1, Distance: 0.1},
			{ID: 4, Distance: 0.2},
			{ID: 2, Distance: 0.3},
			{ID: 7, Distance: 0.3},
			{ID: 5, Distance: 0.4},
		}, hits)

		assert.Equal(t, 1, indicesCallsCount)
		assert.Equal(t, 1, searchKNNCallsCount[1])
		assert.Equal(t, 1, searchKNNCallsCount[2])
		assert.Equal(t, 1, searchKNNCallsCount[4])
	})

	t.Run("Indices request error", func(t *testing.T) {
		t.Parallel()

		tm := time.Date(2000, time.February, 2, 0, 0, 0, 0, tz)

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return nil, errTesting
			},
		}
		c := New(tc, testingConfig)

		hits, err := c.SearchKNN(context.Background(), SearchParams{
			From:              tm,
			To:                tm,
			Vector:            []float32{1, 2, 3},
			DistanceThreshold: 0.5,
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
		assert.Nil(t, hits)
	})

	t.Run("SearchKNN request error", func(t *testing.T) {
		t.Parallel()

		tm := time.Date(2000, time.February, 2, 0, 0, 0, 0, tz)

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return &pb.IndicesReply{Indices: []string{"test_2000-02-01"}}, nil
			},
			searchKNN: func(req *pb.SearchRequest) (*pb.SearchKNNReply, error) {
				return nil, errTesting
			},
		}
		c := New(tc, testingConfig)

		hits, err := c.SearchKNN(context.Background(), SearchParams{
			From:              tm,
			To:                tm,
			Vector:            []float32{1, 2, 3},
			DistanceThreshold: 0.5,
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
		assert.Nil(t, hits)
	})

	t.Run("empty results if no index exists", func(t *testing.T) {
		t.Parallel()

		// This date is "2000-02-01" in UTC
		tm := time.Date(2000, time.February, 2, 0, 0, 0, 0, tz)

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return &pb.IndicesReply{Indices: nil}, nil
			},
		}
		c := New(tc, testingConfig)

		hits, err := c.SearchKNN(context.Background(), SearchParams{
			From:              tm,
			To:                tm,
			Vector:            []float32{1, 2, 3},
			DistanceThreshold: 0.5,
		})
		assert.NoError(t, err)
		assert.Empty(t, hits)
	})
}

func TestClient_IndicesOlderThan(t *testing.T) {
	t.Parallel()

	tz, llErr := time.LoadLocation("Europe/Berlin")
	require.NoError(t, llErr)
	// This date is "2000-02-01" in UTC
	tm := time.Date(2000, time.February, 2, 0, 0, 0, 0, tz)

	t.Run("successful response", func(t *testing.T) {
		t.Parallel()

		indicesCallsCount := 0

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				indicesCallsCount++
				return &pb.IndicesReply{Indices: []string{
					"test_2000-01-30",
					"test_2000-01-31",
					"test_2000-02-01",
					"test_2000-02-02",
				}}, nil
			},
		}
		c := New(tc, testingConfig)

		indices, err := c.IndicesOlderThan(context.Background(), tm)

		assert.NoError(t, err)
		assert.Len(t, indices, 2)
		assert.Contains(t, indices, "test_2000-01-30")
		assert.Contains(t, indices, "test_2000-01-31")

		assert.Equal(t, 1, indicesCallsCount)
	})

	t.Run("Indices response error", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return nil, errTesting
			},
		}
		c := New(tc, testingConfig)

		indices, err := c.IndicesOlderThan(context.Background(), tm)
		assert.Nil(t, indices)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
	})
}

func TestClient_DeleteIndex(t *testing.T) {
	t.Parallel()

	t.Run("successful deletion", func(t *testing.T) {
		t.Parallel()

		deleteIndexCallsCount := 0

		tc := testingClient{
			deleteIndex: func(req *pb.DeleteIndexRequest) error {
				deleteIndexCallsCount++
				assert.Equal(t, "foo", req.GetIndexName())
				return nil
			},
		}
		c := New(tc, testingConfig)

		err := c.DeleteIndex(context.Background(), "foo")
		assert.NoError(t, err)

		assert.Equal(t, 1, deleteIndexCallsCount)
	})

	t.Run("DeleteIndex response error", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			deleteIndex: func(req *pb.DeleteIndexRequest) error {
				return errTesting
			},
		}
		c := New(tc, testingConfig)

		err := c.DeleteIndex(context.Background(), "foo")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
	})
}

func TestClient_FlushAllIndices(t *testing.T) {
	t.Parallel()

	t.Run("successful flush", func(t *testing.T) {
		t.Parallel()

		indicesCallsCount := 0
		flushIndexCallsCount := map[int]int{1: 0, 2: 0}

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				indicesCallsCount++
				return &pb.IndicesReply{Indices: []string{
					"test_2000-02-01",
					"test_2000-02-02",
				}}, nil
			},
			flushIndex: func(req *pb.FlushRequest) error {
				switch req.GetIndexName() {
				case "test_2000-02-01":
					flushIndexCallsCount[1]++
					return nil
				case "test_2000-02-02":
					flushIndexCallsCount[2]++
					return nil
				default:
					assert.FailNow(t, "unexpected FlushIndex call for index %#v", req.GetIndexName())
					return nil
				}
			},
		}
		c := New(tc, testingConfig)

		err := c.FlushAllIndices(context.Background())
		assert.NoError(t, err)

		assert.Equal(t, 1, indicesCallsCount)
		assert.Equal(t, 1, flushIndexCallsCount[1])
		assert.Equal(t, 1, flushIndexCallsCount[2])
	})

	t.Run("Indices request error", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return nil, errTesting
			},
		}
		c := New(tc, testingConfig)

		err := c.FlushAllIndices(context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
	})

	t.Run("FlushIndex request error", func(t *testing.T) {
		t.Parallel()

		tc := testingClient{
			indices: func() (*pb.IndicesReply, error) {
				return &pb.IndicesReply{Indices: []string{"test_2000-02-01"}}, nil
			},
			flushIndex: func(req *pb.FlushRequest) error {
				return errTesting
			},
		}
		c := New(tc, testingConfig)

		err := c.FlushAllIndices(context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTesting)
	})
}

var errTesting = errors.New("an error occurred")

var testingConfig = config.HNSWIndex{
	NamePrefix:     "test_",
	Dim:            3,
	EfConstruction: 200,
	M:              48,
	MaxElements:    100_000,
	Seed:           42,
	SpaceType:      config.HNSWSpaceType(pb.CreateIndexRequest_COSINE),
}

type testingClient struct {
	createIndex        func(req *pb.CreateIndexRequest) error
	deleteIndex        func(req *pb.DeleteIndexRequest) error
	insertVector       func(req *pb.InsertVectorRequest) (*pb.InsertVectorReply, error)
	insertVectorWithID func(req *pb.InsertVectorWithIdRequest) (*pb.InsertVectorWithIdReply, error)
	searchKNN          func(req *pb.SearchRequest) (*pb.SearchKNNReply, error)
	flushIndex         func(req *pb.FlushRequest) error
	indices            func() (*pb.IndicesReply, error)
	setEf              func(req *pb.SetEfRequest) error
}

func (tc testingClient) CreateIndex(
	_ context.Context,
	req *pb.CreateIndexRequest,
	_ ...grpc.CallOption,
) (*emptypb.Empty, error) {
	if tc.createIndex == nil {
		panic("CreateIndex not implemented for testing")
	}
	return nil, tc.createIndex(req)
}

func (tc testingClient) DeleteIndex(
	_ context.Context,
	req *pb.DeleteIndexRequest,
	_ ...grpc.CallOption,
) (*emptypb.Empty, error) {
	if tc.deleteIndex == nil {
		panic("DeleteIndex not implemented for testing")
	}
	return nil, tc.deleteIndex(req)
}

func (tc testingClient) InsertVector(
	_ context.Context,
	req *pb.InsertVectorRequest,
	_ ...grpc.CallOption,
) (*pb.InsertVectorReply, error) {
	if tc.insertVector == nil {
		panic("InsertVector not implemented for testing")
	}
	return tc.insertVector(req)
}

func (tc testingClient) InsertVectors(
	context.Context,
	...grpc.CallOption,
) (pb.Server_InsertVectorsClient, error) {
	panic("InsertVectors not implemented for testing")
}

func (tc testingClient) InsertVectorWithId(
	_ context.Context,
	req *pb.InsertVectorWithIdRequest,
	_ ...grpc.CallOption,
) (*pb.InsertVectorWithIdReply, error) {
	if tc.insertVectorWithID == nil {
		panic("InsertVectorWithId not implemented for testing")
	}
	return tc.insertVectorWithID(req)
}

func (tc testingClient) InsertVectorsWithIds(
	context.Context,
	...grpc.CallOption,
) (pb.Server_InsertVectorsWithIdsClient, error) {
	panic("InsertVectorsWithIds not implemented for testing")
}

func (tc testingClient) SearchKNN(
	_ context.Context,
	req *pb.SearchRequest,
	_ ...grpc.CallOption,
) (*pb.SearchKNNReply, error) {
	if tc.searchKNN == nil {
		panic("SearchKNN not implemented for testing")
	}
	return tc.searchKNN(req)
}

func (tc testingClient) FlushIndex(
	_ context.Context,
	req *pb.FlushRequest,
	_ ...grpc.CallOption,
) (*emptypb.Empty, error) {
	if tc.flushIndex == nil {
		panic("FlushIndex not implemented for testing")
	}
	return nil, tc.flushIndex(req)
}

func (tc testingClient) Indices(
	_ context.Context,
	_ *emptypb.Empty,
	_ ...grpc.CallOption,
) (*pb.IndicesReply, error) {
	if tc.indices == nil {
		panic("Indices not implemented for testing")
	}
	return tc.indices()
}

func (tc testingClient) SetEf(
	_ context.Context,
	req *pb.SetEfRequest,
	_ ...grpc.CallOption,
) (*emptypb.Empty, error) {
	if tc.setEf == nil {
		panic("SetEf not implemented for testing")
	}
	return nil, tc.setEf(req)
}
