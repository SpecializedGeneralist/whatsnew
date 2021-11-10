// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"errors"
	"gorm.io/gorm"
)

// ErrStaleObject is an error returned when attempting to save a stale record.
// A record is stale when it's being saved by another query after instantiation.
// It's part of the optimistic locking mechanism.
var ErrStaleObject = errors.New("stale object error")

// The OptimisticLockModel interface is implemented by GORM models which
// support optimistic locking.
//
// Optimistic locking allows multiple processes to access the same record for
// later updates. This mechanism can be used as a weaker but lighter
// alternative to "pessimistic" locking, that is, using explicit row-level
// locks within transactions.
//
// Optimistic locking is most suitable for operations where minimum conflicts
// are assumed.
//
// It works with models which have an associated version field (corresponding
// to a dedicated column on the related database's table), which is a simple
// monotonically increasing number.
//
// One a model is fetched from the database, the current version must be
// exposed via the method GetVersion. After some changes, the model can
// be saved invoking the function OptimisticSave. This method increases
// the model's version, calling IncrementVersion, and attempts to save the
// record into the database. The operation is successful only if the version
// on the record's database has still the same value of the model's version
// before the increment. If this is not the case, the method fails
// returning the ErrStaleObject error.
type OptimisticLockModel interface {
	GetVersion() uint
	IncrementVersion()
}

// OptimisticSave saves the given model to the database, applyint the
// optimistic locking mechanism.
// It performs a GORM DB.Updates under the hood. If the update fails for any
// reason, the resulting DB.Error is returned unmodified.
// If another process already modified the same record, causing a version
// mismatch, the model is considered "stale", and the error ErrStaleObject is
// returned.
func OptimisticSave(tx *gorm.DB, m OptimisticLockModel) error {
	ver := m.GetVersion()
	m.IncrementVersion()

	ret := tx.Model(m).Select("*").Where("version", ver).Updates(m)
	if ret.Error != nil {
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		return ErrStaleObject
	}
	return nil
}
