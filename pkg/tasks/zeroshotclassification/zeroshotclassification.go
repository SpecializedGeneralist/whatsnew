// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zeroshotclassification

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ZeroShotClassification is the result of a spaGO zero-shot classification.
type ZeroShotClassification struct {
	Class        string                `json:"class"`
	Confidence   float32               `json:"confidence"`
	Distribution []ClassConfidencePair `json:"distribution"`
	Took         int                   `json:"took"`
}

// ClassConfidencePair is a pair of class and confidence.
type ClassConfidencePair struct {
	Class      string  `json:"class"`
	Confidence float32 `json:"confidence"`
}

var _ sql.Scanner = &ZeroShotClassification{}
var _ driver.Valuer = &ZeroShotClassification{}

// Scan implements the Scanner interface.
func (z *ZeroShotClassification) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB ZeroShotClassification value: %#v", value)
	}
	err := json.Unmarshal(bytes, z)
	return err
}

// Value implements the driver Valuer interface.
func (z *ZeroShotClassification) Value() (driver.Value, error) {
	return json.Marshal(z)
}
