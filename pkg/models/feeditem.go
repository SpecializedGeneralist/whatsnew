// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
)

// FeedItem extends a WebResource representing the item of a Feed.
type FeedItem struct {
	Model

	// Association to the Feed this item belongs to.
	FeedID uint `gorm:"not null"`

	// WebResourceID allows the has-one relation with a WebResource.
	WebResourceID uint `gorm:"not null;uniqueIndex;constraint:OnDelete:CASCADE"`

	Title       string `gorm:"not null"`
	Description string `gorm:"not null"`
	Content     string `gorm:"not null"`
	Language    string `gorm:"not null"`
	PublishedAt sql.NullTime
}
