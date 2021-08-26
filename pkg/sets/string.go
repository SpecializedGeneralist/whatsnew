// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sets

type StringSet map[string]emptyStruct

func NewStringSet() StringSet {
	return make(StringSet)
}

func NewStringSetWithSize(size int) StringSet {
	return make(StringSet, size)
}

func (s StringSet) Has(v string) bool {
	_, ok := s[v]
	return ok
}

func (s StringSet) Add(v string) {
	s[v] = emptyStructValue
}

func (s StringSet) Delete(v string) {
	delete(s, v)
}
