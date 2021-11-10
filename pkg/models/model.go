// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import "time"

// Model is the basic struct embedded into all GORM models.
type Model struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`

	// Version for optimistic locking.
	Version uint `gorm:"not null;default:0"`
}

var _ OptimisticLockModel = &Model{}

// GetVersion returns Model.Version, satisfying the
// OptimisticLockModel interface.
func (m Model) GetVersion() uint {
	return m.Version
}

// IncrementVersion increments Model.Version by 1, satisfying the
// OptimisticLockModel interface.
func (m *Model) IncrementVersion() {
	m.Version++
}
