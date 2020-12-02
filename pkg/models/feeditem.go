// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
)

// FeedItem extends a WebResource which is the item of a Feed.
type FeedItem struct {
	Model
	// FeedID is the association to the Feed this item belongs to.
	FeedID uint `gorm:"not null"`
	// WebResourceID allows the has-one relation with a WebResource.
	WebResourceID uint `gorm:"not null;uniqueIndex;constraint:OnDelete:CASCADE"`
	Title         string
	Description   sql.NullString
	Content       sql.NullString
	Language      string `gorm:"not null"`
	PublishedAt   sql.NullTime
}
