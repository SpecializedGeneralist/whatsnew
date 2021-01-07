// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) migrateDB() *cli.Command {
	return &cli.Command{
		Name:  "migrate-db",
		Usage: "Perform automatic database migration",
		Action: func(c *cli.Context) error {
			db, err := models.OpenDB(app.config.DB.DSN, app.newContextLogger(c))
			if err != nil {
				return err
			}
			return models.MigrateDB(db)
		},
	}
}
