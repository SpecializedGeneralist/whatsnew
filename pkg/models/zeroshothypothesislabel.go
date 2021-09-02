// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "gorm.io/gorm"

// ZeroShotHypothesisLabel is one possible label to be replaced in the text of
// a ZeroShotHypothesisTemplate.
type ZeroShotHypothesisLabel struct {
	Model

	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Association to the ZeroShotHypothesisTemplate.
	ZeroShotHypothesisTemplateID uint `gorm:"not null;index;index:idx_hypothesis_id_text,unique"`

	// The system will ignore the labels which are not Enabled.
	Enabled bool `gorm:"not null;index"`

	// Text is the label to be replaced in the hypothesis text.
	Text string `gorm:"not null;index:idx_hypothesis_id_text,unique"`
}
