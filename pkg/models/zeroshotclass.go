// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

// ZeroShotClass is a single class for a WebArticle, predicted with spaGO
// BART zero-shot classification service.
type ZeroShotClass struct {
	Model

	// Association to the WebArticle this class belongs to.
	WebArticleID uint `gorm:"not null;index;index:idx_web_article_id_best,unique,where:best;index:idx_web_article_id_class,unique"`

	Class      string  `gorm:"not null;index;index:idx_web_article_id_class,unique"`
	Confidence float32 `gorm:"not null"`
	Best       bool    `gorm:"not null;index:idx_web_article_id_best,unique,where:best"`
}
