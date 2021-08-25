// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "gorm.io/gorm"

// allModels is the list of all GORM models, used for auto-migration.
var allModels = []interface{}{
	WebResource{},
	WebArticle{},
	Feed{},
	FeedItem{},
	GDELTEvent{},
	TwitterSource{},
	Tweet{},
	PendingJob{},
	ZeroShotClass{},
}

// AutoMigrate performs the automatic migration of all GORM models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(allModels...)
}
