// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"regexp"
)

// Regexp is a compiled regular expression that can be stored to database
// as string value.
type Regexp struct {
	*regexp.Regexp
}

var _ sql.Scanner = &Regexp{}
var _ driver.Valuer = Regexp{}
var _ schema.GormDataTypeInterface = Regexp{}
var _ migrator.GormDataTypeInterface = Regexp{}

// Scan assigns a value from a database driver.
// It expects a string value, which is compiled to a regexnilp.Regexp.
func (r *Regexp) Scan(value interface{}) error {
	expr, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan Regexp %#v", value)
	}

	var err error
	r.Regexp, err = regexp.Compile(expr)
	if err != nil {
		return fmt.Errorf("cannot scan Regexp %#v: %w", value, err)
	}
	return nil
}

// Value returns a driver Value.
func (r Regexp) Value() (driver.Value, error) {
	if r.Regexp == nil {
		return nil, fmt.Errorf("a compiled regular expression is missing")
	}
	return r.String(), nil
}

// GormDataType returns the data type for GORM integration.
func (r Regexp) GormDataType() string {
	return "text"
}

// GormDBDataType returns the data type for GORM integration.
func (r Regexp) GormDBDataType(*gorm.DB, *schema.Field) string {
	return "text"
}
