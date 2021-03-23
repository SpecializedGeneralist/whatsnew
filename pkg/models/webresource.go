// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"gorm.io/gorm"
)

// WebResource represents a web resource, usually a web page, accessible
// via a URL.
type WebResource struct {
	Model
	// URL is the unique URL of the web resource.
	URL string `gorm:"not null;uniqueIndex"`
	// WebArticle represents the scraped content of this WebResource.
	WebArticle WebArticle
	// FeedItem allows the has-one relation with a models.FeedItem.
	FeedItem FeedItem
	// GDELTEvent allows the has-one relation with a models.GDELTEvent.
	GDELTEvent GDELTEvent
	// Tweet allows the has-one relation with a models.Tweet.
	Tweet Tweet
}

// FindWebResourceByURL returns the web resource associated to a url.
// It returns nil if the item is not found.
func FindWebResourceByURL(tx *gorm.DB, url string) (*WebResource, error) {
	var webResources []*WebResource
	result := tx.Find(&webResources, "url = ?", url)
	if result.Error != nil {
		return nil, fmt.Errorf("find web resource by URL %#v: %v", url, result.Error)
	}
	if len(webResources) == 0 {
		return nil, nil
	}
	return webResources[0], nil
}
