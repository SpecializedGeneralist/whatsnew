// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "gorm.io/gorm"

// ZeroShotHypothesisTemplate represents the template for on hypothesis
// used for BART zero-shot classification of WebArticles.
type ZeroShotHypothesisTemplate struct {
	Model

	DeletedAt gorm.DeletedAt `gorm:"index"`

	// The system will ignore the templates which are not Enabled.
	Enabled bool `gorm:"not null;index"`

	// Text is the hypothesis. It MUST contain one character sequence "{}" to
	// indicate the point where each related label will be placed.
	Text string `gorm:"not null"`

	// MultiClass indicates whether the zero-shot classification is multi-class
	// (true) or single-class (false).
	MultiClass bool `gorm:"not null"`

	// Labels are the possible items to be replaced in the Text.
	Labels []ZeroShotHypothesisLabel `gorm:"constraint:OnDelete:CASCADE"`
}
