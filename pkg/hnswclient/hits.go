// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hnswclient

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
