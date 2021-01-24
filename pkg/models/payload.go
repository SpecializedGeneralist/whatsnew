// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Payload is a generic payload for a WebArticle.
type Payload map[string]interface{}

var _ sql.Scanner = &Payload{}
var _ driver.Valuer = &Payload{}

// Set sets a key/value pair in the Payload.
// It initializes the payload, if necessary. If a value already exists for
// the given key, it is overwritten.
func (p *Payload) Set(key string, value interface{}) {
	if *p == nil {
		*p = make(Payload, 1)
	}
	(*p)[key] = value
}

// Get gets a value from the given key.
// If multiple keys are given, the method attempts to match a series of nested maps.
// If a value is not found, or the search in nested maps fails, the returned value is nil,
// and the flag is false.
func (p *Payload) Get(key string, nestedKeys ...string) (interface{}, bool) {
	if *p == nil {
		return nil, false
	}
	value, ok := (*p)[key]
	if !ok {
		return nil, false
	}
	for _, k := range nestedKeys {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil, false
		}
		value = m[k]
	}
	return value, true
}

// GetString gets the value with Get and enforces it to be a string.
// If no value is found, or the value is not a string, the method returns
// an empty string and a false flag.
func (p *Payload) GetString(key string, nestedKeys ...string) (string, bool) {
	v, ok := p.Get(key, nestedKeys...)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}

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
