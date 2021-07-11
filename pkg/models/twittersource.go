// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "database/sql"

// TwitterSourceType acts as an enumeration type to identify different kind
// of Twitter sources.
type TwitterSourceType string

const (
	// UserTwitterSource identifies a Twitter source linked to a user profile.
	UserTwitterSource TwitterSourceType = "user"

	// SearchTwitterSource identifies a Twitter source linked to a terms search.
	SearchTwitterSource TwitterSourceType = "search"
)

// TwitterSource represents the source of Tweets / WebResources.
type TwitterSource struct {
	Model

	Type TwitterSourceType `gorm:"not null;index:idx_twitter_sources_type_text,unique"`

	// Text is either a username or a search term, depending on the Type.
	Text string `gorm:"not null;index:idx_twitter_sources_type_text,unique"`

	// The system will look for new tweets from this source only when it is
	// Enabled. Otherwise, the twitter source is simply ignored.
	Enabled bool `gorm:"not null;index;default:true"`

	// The date and time when this source was last visited to successfully
	// retrieve its content (tweets), store it, and schedule further
	// processing jobs.
	LastRetrievedAt sql.NullTime `gorm:"index"`

	// Counter of consecutive fetching failures.
	FailuresCount int `gorm:"not null;default:0"`

	// When FailuresCount is not 0, this field should contain the error message
	// that caused the last failure. It is mostly useful for manual inspection.
	LastError sql.NullString

	// Tweets is the has-many relation with Tweet models.
	Tweets []Tweet `gorm:"constraint:OnDelete:CASCADE"`
}
