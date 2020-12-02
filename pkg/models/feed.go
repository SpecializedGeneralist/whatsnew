// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
)

// Feed is a model representing an RSS or Atom feed.
type Feed struct {
	Model
	DeletedAt gorm.DeletedAt `gorm:"index"`
	// URL is the unique URL of the feed.
	URL string `gorm:"not null;uniqueIndex"`
	// LastRetrievedAt is the date and time when this feed was last visited
	// to successfully retrieve and process its content.
	LastRetrievedAt sql.NullTime `gorm:"index"`
	FailuresCount   int          `gorm:"not null"`
	// FeedItems is the has-many relation with FeedItem models.
	FeedItems []FeedItem `gorm:"constraint:OnDelete:CASCADE"`
}

// FindFeed returns the feed by its id.
// It returns nil if the item is not found.
func FindFeed(db *gorm.DB, id uint) (*Feed, error) {
	feed := &Feed{}
	result := db.First(feed, id)
	if result.Error != nil {
		return nil, fmt.Errorf("find Feed %d: %v", id, result.Error)
	}
	return feed, nil
}
