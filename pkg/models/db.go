// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"
	"github.com/nlpodyssey/whatsnew/pkg/gormlogger"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OpenDB initializes a database session.
func OpenDB(dsn string, logger zerolog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.New(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("open database session: %v", err)
	}
	return db, nil
}

// MigrateDB performs an automatic migration of all GORM models.
func MigrateDB(db *gorm.DB) error {
	allModels := GetAllModels()
	return db.AutoMigrate(allModels...)
}
