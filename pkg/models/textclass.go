// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

// TextClass is a single classification result class for a WebArticle,
// predicted with a generic text classifier.
type TextClass struct {
	Model

	// Association to the WebArticle this class belongs to.
	WebArticleID uint `gorm:"not null;index"`

	Type       string  `gorm:"not null;index"`
	Label      string  `gorm:"not null;index"`
	Confidence float32 `gorm:"not null"`
}
