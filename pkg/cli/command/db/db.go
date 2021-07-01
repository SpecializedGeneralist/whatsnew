// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
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
	panic("not implemented")
}
