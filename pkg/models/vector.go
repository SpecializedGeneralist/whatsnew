// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "github.com/jackc/pgtype"

// Vector is a vector representation of a WebArticle.
type Vector struct {
	Model

	// Association to the WebArticle this vector belongs to.
	WebArticleID uint `gorm:"not null;uniqueIndex"`

	Data *pgtype.Float4Array `gorm:"type:float4[];not null"`
}
