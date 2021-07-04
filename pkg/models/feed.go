// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "database/sql"

// Feed is a model representing an RSS or Atom feed.
type Feed struct {
	Model

	// The unique URL of the feed.
	URL string `gorm:"not null;uniqueIndex"`

	// The system will look for new feed items from this feed only when it is
	// Enabled. Otherwise, the feed is simply ignored.
	Enabled bool `gorm:"not null;index;default:true"`

	// The date and time when this feed was last visited to successfully
	// retrieve its content (feed items), store it, and schedule further
	// processing jobs.
	LastRetrievedAt sql.NullTime `gorm:"index"`

	// Counter of consecutive fetching failures.
	FailuresCount int `gorm:"not null;default:0"`

	// A Feed has many models.FeedItem models.
	FeedItems []FeedItem `gorm:"constraint:OnDelete:CASCADE"`
}
