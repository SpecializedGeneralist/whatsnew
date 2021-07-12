// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

// WebResource represents a web resource, usually a web page, accessible
// via an URL.
type WebResource struct {
	Model

	// The unique URL of the web resource.
	URL string `gorm:"not null;uniqueIndex"`

	// A WebArticle extends the WebResource with the scraped content.
	WebArticle *WebArticle `gorm:"constraint:OnDelete:CASCADE"`

	// FeedItem allows the has-one relation with a models.FeedItem.
	FeedItem *FeedItem `gorm:"constraint:OnDelete:CASCADE"`

	// GDELTEvent allows the has-one relation with a models.GDELTEvent.
	GDELTEvent *GDELTEvent `gorm:"constraint:OnDelete:CASCADE"`

	// Tweet allows the has-one relation with a models.Tweet.
	Tweet *Tweet `gorm:"constraint:OnDelete:CASCADE"`
}
