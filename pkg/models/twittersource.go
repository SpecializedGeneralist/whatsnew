// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
)

const (
	UserTwitterSource   = "user"
	SearchTwitterSource = "search"
)

// TwitterSource is a model representing a Twitter user or search phrase.
type TwitterSource struct {
	Model
	// Type is either "user" or "search".
	Type string `gorm:"not null;index:type_value,unique"`
	// Value is a username or a search term, depending on the Type.
	Value string `gorm:"not null;index:type_value,unique"`
	// LastRetrievedAt is the date and time when this Twitter source was last
	// visited to successfully retrieve and process its content.
	LastRetrievedAt sql.NullTime `gorm:"index"`
	// Tweet is the has-many relation with Tweet models.
	Tweets []Tweet `gorm:"constraint:OnDelete:CASCADE"`
}

// FindTwitterSource returns a TwitterSource by its id.
// It returns nil if the item is not found.
func FindTwitterSource(db *gorm.DB, id uint) (*TwitterSource, error) {
	ts := &TwitterSource{}
	result := db.First(ts, id)
	if result.Error != nil {
		return nil, fmt.Errorf("find TwitterSource %d: %v", id, result.Error)
	}
	return ts, nil
}
