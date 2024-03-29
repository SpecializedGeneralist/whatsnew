// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

// ExtractedInfo is a single result from the information extraction task
// performed on a WebArticle.
type ExtractedInfo struct {
	Model

	// Association to the WebArticle.
	WebArticleID uint `gorm:"not null;index;index:idx_web_article_id_info_extraction_rule_id,unique"`

	// Association to the InfoExtractionRule.
	InfoExtractionRuleID uint `gorm:"not null;index;index:idx_web_article_id_info_extraction_rule_id,unique"`

	Text       string  `gorm:"not null"`
	Confidence float32 `gorm:"not null"`
}
