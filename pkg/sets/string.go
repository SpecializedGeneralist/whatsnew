// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sets

// StringSet is a set of unique strings.
type StringSet map[string]emptyStruct

// NewStringSet creates a new StringSet, pre-allocating a small starting size
// of memory.
func NewStringSet() StringSet {
	return make(StringSet)
}

// NewStringSetWithSize creates a new StringSet, pre-allocating the specified
// starting size of memory.
func NewStringSetWithSize(size int) StringSet {
	return make(StringSet, size)
}

// Has reports whether v is included in the set.
func (s StringSet) Has(v string) bool {
	_, ok := s[v]
	return ok
}

// Add adds v to the set.
func (s StringSet) Add(v string) {
	s[v] = emptyStructValue
}

// Delete deletes v from the set.
func (s StringSet) Delete(v string) {
	delete(s, v)
}
