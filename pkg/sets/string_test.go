// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sets

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStringSet(t *testing.T) {
	t.Parallel()
	s := NewStringSet()
	assert.Empty(t, s)
}

func TestNewStringSetWithSize(t *testing.T) {
	t.Parallel()
	s := NewStringSetWithSize(10)
	assert.Empty(t, s)
}

func TestNewStringSetWithElements(t *testing.T) {
	t.Parallel()
	s := NewStringSetWithElements("foo", "bar")
	assert.Len(t, s, 2)
	assert.True(t, s.Has("foo"))
	assert.True(t, s.Has("bar"))
	assert.False(t, s.Has("baz"))
}

func TestStringSet(t *testing.T) {
	t.Parallel()
	s := NewStringSetWithSize(10)

	assert.False(t, s.Has("foo"))
	assert.False(t, s.Has("bar"))
	assert.Empty(t, s)

	s.Add("foo")

	assert.Len(t, s, 1)
	assert.True(t, s.Has("foo"))
	assert.False(t, s.Has("bar"))

	s.Add("bar")

	assert.Len(t, s, 2)
	assert.True(t, s.Has("foo"))
	assert.True(t, s.Has("bar"))

	s.Delete("foo")

	assert.Len(t, s, 1)
	assert.False(t, s.Has("foo"))
	assert.True(t, s.Has("bar"))

	s.Delete("bar")

	assert.Empty(t, s)
	assert.False(t, s.Has("foo"))
	assert.False(t, s.Has("bar"))
}

func TestStringSet_AddMany(t *testing.T) {
	t.Parallel()
	s := NewStringSet()
	s.AddMany("foo", "bar")
	assert.Len(t, s, 2)
	assert.True(t, s.Has("foo"))
	assert.True(t, s.Has("bar"))
	assert.False(t, s.Has("baz"))
}
