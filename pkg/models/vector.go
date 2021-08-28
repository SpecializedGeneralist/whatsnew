// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"github.com/jackc/pgtype"
)

// Vector is a vector representation of a WebArticle.
type Vector struct {
	Model

	// Association to the WebArticle this vector belongs to.
	WebArticleID uint `gorm:"not null;uniqueIndex"`

	Data *pgtype.Float4Array `gorm:"type:float4[];not null"`
}

func (v Vector) DataAsFloat32Slice() ([]float32, error) {
	var vec []float32
	err := v.Data.AssignTo(&vec)
	if err != nil {
		return nil, fmt.Errorf("error converting Vector.Data to []float32: %w", err)
	}
	return vec, nil
}
