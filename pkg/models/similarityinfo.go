// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

// SimilarityInfo provides information about similarity between WebArticles.
//
// If a WebArticle "B" is considered to be a similar (or duplicate) of
// WebArticle "A"then "A" is considered the "parent" of "B".
//
// If a WebArticle has no related SimilarityInfo, it means that the similarity
// detection task was not (yet) performed, or an error occurred.
//
// If a WebArticle has a SimilarityInfo attached with no Parent (and no
// Distance), it means that the similarity detection task was successfully
// performed and no prior similar entities were found.
type SimilarityInfo struct {
	Model

	// Association to the WebArticle this info belongs to.
	WebArticleID uint `gorm:"not null;uniqueIndex"`

	ParentID *uint `gorm:"index"`
	Parent   *WebArticle

	Distance *float32
}
