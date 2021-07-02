// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"gorm.io/gorm"
)

// CmdDB implements the command "whatsnew db".
var CmdDB = &command.Command{
	Name:      "db",
	UsageLine: "db (create | migrate | delete)",
	Short:     "perform basic operations on the database",
	Long: `
The "db" command performs basic operations on the database.

You must specify a sub-command, which identifies the specific operation to
perform. The supported operations are:

	create
		Create the new database and initializes the schema according to
		the application models.

	migrate
		Perform automatic schema migration on an existing database.
		It corresponds to running GORM "db.AutoMigrate" function for
		all application models. See https://gorm.io/docs/migration.html for
		more details.

	drop
		Drop the database.
`,
	Run: Run,
}

// Run runs the command "whatsnew db".
func Run(conf *config.Config, args []string) error {
	if len(args) != 1 {
		return command.ErrInvalidArguments
	}
	fn, ok := operations[args[0]]
	if !ok {
		return command.ErrInvalidArguments
	}
	return fn(conf)
}

type opFn func(conf *config.Config) error

var operations = map[string]opFn{
	"create":  runCreate,
	"migrate": runMigrate,
	"drop":    runDrop,
}

func runCreate(conf *config.Config) (err error) {
	db, err := database.OpenDBWithoutDBName(conf.DB)
	if err != nil {
		return err
	}
	defer func() {
		if e := database.CloseDB(db); e != nil && err == nil {
			err = e
		}
	}()
	err = createDB(db, conf.DB.DBName)
	if err != nil {
		return err
	}
	err = database.CloseDB(db)
	if err != nil {
		return err
	}
	return runMigrate(conf)
}

func runMigrate(conf *config.Config) (err error) {
	db, err := database.OpenDB(conf.DB)
	if err != nil {
		return err
	}
	defer func() {
		if e := database.CloseDB(db); e != nil && err == nil {
			err = e
		}
	}()
	return models.AutoMigrate(db)
}

func runDrop(conf *config.Config) (err error) {
	db, err := database.OpenDBWithoutDBName(conf.DB)
	if err != nil {
		return err
	}
	defer func() {
		if e := database.CloseDB(db); e != nil && err == nil {
			err = e
		}
	}()
	return dropDB(db, conf.DB.DBName)
}

func createDB(db *gorm.DB, name string) error {
	quotedDBName, err := quoteIdent(db, name)
	if err != nil {
		return err
	}
	res := db.Exec(fmt.Sprintf("CREATE DATABASE %s", quotedDBName))
	if res.Error != nil {
		return fmt.Errorf("error creating database: %w", res.Error)
	}
	return nil
}

func dropDB(db *gorm.DB, name string) error {
	quotedDBName, err := quoteIdent(db, name)
	if err != nil {
		return err
	}
	res := db.Exec(fmt.Sprintf("DROP DATABASE %s", quotedDBName))
	if res.Error != nil {
		return fmt.Errorf("error dropping database: %w", res.Error)
	}
	return nil
}

func quoteIdent(db *gorm.DB, s string) (string, error) {
	res := db.Raw("SELECT quote_ident(?)", s)
	if res.Error != nil {
		return "", fmt.Errorf("error escaping database name: %w", res.Error)
	}
	var quoted string
	row := res.Row()
	err := row.Scan(&quoted)
	if err != nil {
		return "", fmt.Errorf("error escaping database name: %w", res.Error)
	}
	return quoted, nil
}
