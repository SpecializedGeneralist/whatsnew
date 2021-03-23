// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"time"
)

// Tweet extends a WebResource which is the item of a TwitterSource.
type Tweet struct {
	Model
	// TwitterSourceID is the association to the TwitterSource this item belongs to.
	TwitterSourceID uint `gorm:"not null"`
	// WebResourceID allows the has-one relation with a WebResource.
	WebResourceID uint `gorm:"not null;uniqueIndex;constraint:OnDelete:CASCADE"`

	UpstreamID  string    `gorm:"not null;uniqueIndex"`
	Text        string    `gorm:"not null"`
	PublishedAt time.Time `gorm:"not null"`
	Username    string    `gorm:"not null;index"`
	UserID      string    `gorm:"not null;index"`
}
