// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
	"time"
)

// WebArticle represents the scraped content of a WebResource.
type WebArticle struct {
	Model
	// WebResourceID allows the has-one relation with a WebResource.
	WebResourceID         uint   `gorm:"not null;uniqueIndex;constraint:OnDelete:CASCADE"`
	Title                 string `gorm:"not null;index"`
	TitleUnmodified       string
	CleanedText           string
	CanonicalLink         string
	TopImage              string
	FinalURL              string
	ScrapedPublishDate    sql.NullTime
	Language              string    `gorm:"not null"`
	PublishDate           time.Time `gorm:"not null"`
	RelatedToWebArticleID *uint
	RelatedToWebArticle   *WebArticle
	RelatedScore          sql.NullFloat64
	Payload               Payload      `gorm:"type:JSONB"`
	Vector                pgtype.Bytea `gorm:"type:bytea"`
}

// FindWebArticle returns the web article by its id.
// It returns nil if the item is not found.
func FindWebArticle(tx *gorm.DB, id uint) (*WebArticle, error) {
	webArticle := &WebArticle{}
	result := tx.First(webArticle, id)
	if result.Error != nil {
		return nil, fmt.Errorf("find WebArticle %d: %v", id, result.Error)
	}
	if webArticle.Vector.Status == pgtype.Undefined {
		// FIXME: if the value is null, the status is Undefined instead of null
		//        so saving the article back would fail.
		webArticle.Vector.Status = pgtype.Null
	}
	return webArticle, nil
}

// Payload is a generic payload for a WebArticle.
type Payload map[string]interface{}

var _ sql.Scanner = &Payload{}
var _ driver.Valuer = &Payload{}

func (p *Payload) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB Payload value of type %T: %#v", src, src)
	}
	if len(bytes) == 0 {
		*p = make(Payload, 0)
		return nil
	}
	err := json.Unmarshal(bytes, p)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSONB Payload %#v: %w", string(bytes), err)
	}
	return nil
}

func (p *Payload) Value() (driver.Value, error) {
	if len(*p) == 0 {
		return nil, nil
	}
	v, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Payload %#v: %w", *p, err)
	}
	return v, nil
}
