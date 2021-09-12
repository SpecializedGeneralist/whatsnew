// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cliff

// ParsedText is the result of parsing a text.
type ParsedText struct {
	Results ParsedTextResults `json:"results"`
}

// ParsedTextResults contains the data extracted from a text.
type ParsedTextResults struct {
	Places Places `json:"places"`
}

// The Places extracted from a text.
type Places struct {
	Focus Focus `json:"focus"`
}

// The Focus extracted entities.
type Focus struct {
	Cities    []Location `json:"cities"`
	Countries []Location `json:"countries"`
	States    []Location `json:"states"`
}

// AllLocations returns the union of Cities, Countries and States.
func (f *Focus) AllLocations() []Location {
	all := make([]Location, 0, len(f.Cities)+len(f.Countries)+len(f.States))
	all = append(all, f.Cities...)
	all = append(all, f.Countries...)
	all = append(all, f.States...)
	return all
}

// A Location extracted from text.
type Location struct {
	CountryCode string  `json:"countryCode"`
	Score       float64 `json:"score"`
}
