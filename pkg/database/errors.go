// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"github.com/jackc/pgconn"
)

// IsUniqueViolationError reports whether err is (or wraps at any level) a
// pgconn.PgError with SQLSTATE corresponding to Postgres "unique_violation"
// (error code "23505").
func IsUniqueViolationError(err error) bool {
	if err == nil {
		return false
	}
	var target *pgconn.PgError
	if errors.As(err, &target) {
		return target.SQLState() == "23505"
	}
	return false
}
