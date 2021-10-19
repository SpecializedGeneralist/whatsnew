// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hnswclient

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestHit_LessThan(t *testing.T) {
	t.Parallel()

	t.Run("it compares the distance", func(t *testing.T) {
		t.Parallel()
		a := Hit{ID: 1, Distance: 1}
		b := Hit{ID: 3, Distance: 2}
		c := Hit{ID: 2, Distance: 3}
		assert.True(t, a.LessThan(b))
		assert.True(t, a.LessThan(c))
		assert.True(t, b.LessThan(c))
		assert.False(t, c.LessThan(b))
		assert.False(t, c.LessThan(a))
		assert.False(t, b.LessThan(a))
	})

	t.Run("it compares the ID if distance is identical", func(t *testing.T) {
		t.Parallel()
		a := Hit{ID: 1, Distance: 42}
		b := Hit{ID: 2, Distance: 42}
		c := Hit{ID: 3, Distance: 42}
		assert.True(t, a.LessThan(b))
		assert.True(t, a.LessThan(c))
		assert.True(t, b.LessThan(c))
		assert.False(t, c.LessThan(b))
		assert.False(t, c.LessThan(a))
		assert.False(t, b.LessThan(a))
	})
}

func TestHits_Len(t *testing.T) {
	t.Parallel()
	hits := make(Hits, 0)
	assert.Equal(t, 0, hits.Len())
	hits = append(hits, Hit{})
	assert.Equal(t, 1, hits.Len())
	hits = append(hits, Hit{})
	assert.Equal(t, 2, hits.Len())
}

func TestHits_Less(t *testing.T) {
	t.Parallel()

	t.Run("it compares the distance", func(t *testing.T) {
		t.Parallel()
		hits := Hits{
			Hit{ID: 1, Distance: 1},
			Hit{ID: 3, Distance: 2},
			Hit{ID: 2, Distance: 3},
		}
		assert.True(t, hits.Less(0, 1))
		assert.True(t, hits.Less(0, 2))
		assert.True(t, hits.Less(1, 2))
		assert.False(t, hits.Less(2, 1))
		assert.False(t, hits.Less(2, 0))
		assert.False(t, hits.Less(1, 0))
	})

	t.Run("it compares the ID if distance is identical", func(t *testing.T) {
		t.Parallel()
		hits := Hits{
			Hit{ID: 1, Distance: 42},
			Hit{ID: 2, Distance: 42},
			Hit{ID: 3, Distance: 42},
		}
		assert.True(t, hits.Less(0, 1))
		assert.True(t, hits.Less(0, 2))
		assert.True(t, hits.Less(1, 2))
		assert.False(t, hits.Less(2, 1))
		assert.False(t, hits.Less(2, 0))
		assert.False(t, hits.Less(1, 0))
	})
}

func TestHits_Swap(t *testing.T) {
	t.Parallel()

	hits := Hits{
		Hit{ID: 1, Distance: 10},
		Hit{ID: 2, Distance: 20},
		Hit{ID: 3, Distance: 30},
	}

	hits.Swap(0, 1)
	assert.Equal(t, Hits{
		Hit{ID: 2, Distance: 20},
		Hit{ID: 1, Distance: 10},
		Hit{ID: 3, Distance: 30},
	}, hits)

	hits.Swap(1, 2)
	assert.Equal(t, Hits{
		Hit{ID: 2, Distance: 20},
		Hit{ID: 3, Distance: 30},
		Hit{ID: 1, Distance: 10},
	}, hits)
}

func TestHits(t *testing.T) {
	t.Parallel()

	t.Run("hits are sorted correctly", func(t *testing.T) {
		t.Parallel()
		hits := Hits{
			Hit{ID: 4, Distance: 40},
			Hit{ID: 1, Distance: 10},
			Hit{ID: 3, Distance: 23},
			Hit{ID: 1, Distance: 11},
			Hit{ID: 2, Distance: 23},
		}
		sort.Sort(hits)
		assert.Equal(t, Hits{
			Hit{ID: 1, Distance: 10},
			Hit{ID: 1, Distance: 11},
			Hit{ID: 2, Distance: 23},
			Hit{ID: 3, Distance: 23},
			Hit{ID: 4, Distance: 40},
		}, hits)
	})
}
