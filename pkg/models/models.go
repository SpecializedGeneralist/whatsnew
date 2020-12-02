// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

// GetAllModels returns all GORM models, that is, a slice of pointers to
// zero-valued structs.
func GetAllModels() []interface{} {
	return []interface{}{
		&WebResource{},
		&Feed{},
		&WebArticle{},
		&FeedItem{},
		&GDELTEvent{},
	}
}
