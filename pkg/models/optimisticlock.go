// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"errors"
	"gorm.io/gorm"
)

type OptimisticLockModel interface {
	GetVersion() uint
	IncrementVersion()
}

func OptimisticSave(tx *gorm.DB, m OptimisticLockModel) error {
	ver := m.GetVersion()
	m.IncrementVersion()

	ret := tx.Model(m).Where("version", ver).UpdateColumns(m)
	if ret.Error != nil {
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		return errors.New("optimistic-lock saving failed")
	}
	return nil
}
