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

	// FeedItem allows the has-one relation with a models.FeedItem.
	FeedItem FeedItem
}
