// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// WebArticle represents the scraped content of a WebResource.
type WebArticle struct {
	Model
	// WebResourceID allows the has-one relation with a WebResource.
	WebResourceID      uint   `gorm:"not null;uniqueIndex;constraint:OnDelete:CASCADE"`
	Title              string `gorm:"not null;index"`
	TitleUnmodified    string
	CleanedText        string
	CanonicalLink      string
	TopImage           string
	FinalURL           string
	ScrapedPublishDate sql.NullTime
	Language           string    `gorm:"not null"`
	PublishDate        time.Time `gorm:"not null"`
}

// FindWebArticle returns the web article by its id.
// It returns nil if the item is not found.
func FindWebArticle(tx *gorm.DB, id uint) (*WebArticle, error) {
	webArticle := &WebArticle{}
	result := tx.First(webArticle, id)
	if result.Error != nil {
		return nil, fmt.Errorf("find WebArticle %d: %v", id, result.Error)
	}
	return webArticle, nil
}
