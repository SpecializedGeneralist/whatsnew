// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models/types"
	"gorm.io/gorm"
)

// InfoExtractionRule is a single configuration item for the information
// extraction task.
type InfoExtractionRule struct {
	Model

	DeletedAt gorm.DeletedAt `gorm:"index"`

	Label        string       `gorm:"not null;uniqueIndex"`
	Question     string       `gorm:"not null"`
	AnswerRegexp types.Regexp `gorm:"not null"`
	Threshold    float32      `gorm:"not null"`
	Enabled      bool         `gorm:"not null;index"`
}
