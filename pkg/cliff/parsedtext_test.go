// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cliff

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFocus_AllLocations(t *testing.T) {
	f := Focus{}

	assert.Empty(t, f.AllLocations())

	f.Cities = []Location{
		{CountryCode: "AD", Score: 0.1},
		{CountryCode: "AE", Score: 0.2},
	}
	f.Countries = []Location{
		{CountryCode: "BA", Score: 0.3},
		{CountryCode: "BB", Score: 0.4},
	}
	f.States = []Location{
		{CountryCode: "CC", Score: 0.5},
		{CountryCode: "CD", Score: 0.6},
	}

	expected := []Location{
		{CountryCode: "AD", Score: 0.1},
		{CountryCode: "AE", Score: 0.2},
		{CountryCode: "BA", Score: 0.3},
		{CountryCode: "BB", Score: 0.4},
		{CountryCode: "CC", Score: 0.5},
		{CountryCode: "CD", Score: 0.6},
	}
	assert.Equal(t, expected, f.AllLocations())
}
