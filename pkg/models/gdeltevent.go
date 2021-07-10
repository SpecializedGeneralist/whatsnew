// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"github.com/jackc/pgtype"
	"time"
)

// GDELTEvent represents a GDELT Event, and it extends a WebResource which
// contains the URL of the first recognized news report for this event.
type GDELTEvent struct {
	Model

	// WebResourceID allows the has-one relation with a WebResource.
	WebResourceID uint `gorm:"not null;uniqueIndex"`

	// GlobalEventID is the globally unique identifier in GDELT master dataset.
	GlobalEventID uint `gorm:"not null;uniqueIndex"`

	// DateAdded is the date the event was added to the master database.
	DateAdded time.Time `gorm:"not null"`

	// LocationType specifies the geographic resolution of the match type.
	LocationType sql.NullString

	// LocationName is the full human-readable name of the matched location.
	LocationName sql.NullString

	// CountryCode is the ISO 3166-1 alpha2 country code for the location.
	CountryCode sql.NullString

	// Coordinates provides the centroid Longitude (X) and Latitude (Y) of
	// the landmark for mapping.
	Coordinates pgtype.Point `gorm:"type:point"`

	// EventCategories provides one or more CAMEO event codes at different
	// levels.
	EventCategories pgtype.TextArray `gorm:"type:text[];not null"`
}
