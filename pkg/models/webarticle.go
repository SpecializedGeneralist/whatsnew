// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"time"
)

// WebArticle represents the scraped content of a WebResource.
type WebArticle struct {
	Model

	// WebResourceID allows the has-one relation with a WebResource.
	WebResourceID uint `gorm:"not null;uniqueIndex"`

	Title              string `gorm:"not null;index"`
	TopImage           sql.NullString
	ScrapedPublishDate sql.NullTime
	Language           string    `gorm:"not null"`
	PublishDate        time.Time `gorm:"not null"`

	TranslatedTitle     sql.NullString
	TranslationLanguage sql.NullString

	// A WebArticle has many models.ZeroShotClass models.
	ZeroShotClasses []ZeroShotClass `gorm:"constraint:OnDelete:CASCADE"`

	// A WebArticle has many models.TextClass models.
	TextClasses []TextClass `gorm:"constraint:OnDelete:CASCADE"`

	// A WebArticle has many models.ExtractedInfo models.
	ExtractedInfos []ExtractedInfo `gorm:"constraint:OnDelete:CASCADE"`

	// A WebArticle has one Vector.
	Vector *Vector `gorm:"constraint:OnDelete:CASCADE"`

	// A WebArticle has one SimilarityInfo.
	SimilarityInfo *SimilarityInfo `gorm:"constraint:OnDelete:CASCADE"`
}
