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

type OptimisticLockModel interface {
	GetVersion() uint
	IncrementVersion()
}

func OptimisticSave(tx *gorm.DB, m OptimisticLockModel) error {
	ver := m.GetVersion()
	m.IncrementVersion()

	ret := tx.Model(m).Where("version", ver).Updates(m)
	if ret.Error != nil {
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		return ErrStaleObject
	}
	return nil
}
