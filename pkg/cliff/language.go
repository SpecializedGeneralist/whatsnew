// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cliff

// A Language supported by CLIFF.
type Language string

const (
	// German language.
	German Language = "DE"
	// Spanish language.
	Spanish Language = "ES"
	// English language.
	English Language = "EN"
)

// String returns the CLIFF string representation of the language.
func (l Language) String() string {
	return string(l)
}
