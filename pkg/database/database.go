// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package database provides useful functions for database operations and
// configuration.
package database

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// OpenDB initializes a database session, connecting to the specific database
// from configuration.
func OpenDB(conf config.DB) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s dbname=%s", conf.DSN, conf.DBName)
	return openDB(dsn, conf.LogLevel)
}

// OpenDBWithoutDBName initializes a database session ignoring config.DB.DBName.
//
// Among others, it can be useful for special operations such as creating or
// dropping a database.
func OpenDBWithoutDBName(conf config.DB) (*gorm.DB, error) {
	return openDB(conf.DSN, conf.LogLevel)
}

// CloseDB closes the underlying sql.DB from a gorm.DB.
func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("error getting sql.DB from gorm.DB: %w", err)
	}
	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("error closing database: %w", err)
	}
	return nil
}

func openDB(dsn string, logLevel config.DBLogLevel) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), gormConfig(logLevel))
	if err != nil {
		return nil, fmt.Errorf("error opening database session: %w", err)
	}
	return db, nil
}

func gormConfig(logLevel config.DBLogLevel) *gorm.Config {
	return &gorm.Config{
		Logger: NewGORMLogger().LogMode(gormlogger.LogLevel(logLevel)),
	}
}
