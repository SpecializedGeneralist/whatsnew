// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) createDB() *cli.Command {
	return &cli.Command{
		Name:  "create-db",
		Usage: "Perform automatic database creation",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Required: true,
				Usage:    "the name of the database to create",
			},
		},
		Action: func(c *cli.Context) error {
			db, err := models.OpenDB(app.config.DB.DSN, app.newContextLogger(c))
			if err != nil {
				return err
			}

			// `db.Exec` argument interpolation does not work properly, causing the
			// command to always fail. So we build the statement with a humble
			// `fmt.Sprintf`, just providing minimal validation.
			dbName := c.String("name")
			if err = validateDatabaseName(dbName); err != nil {
				return err
			}

			result := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
			if result.Error != nil {
				return result.Error
			}
			fmt.Printf("Done!")
			return nil
		},
	}
}

func validateDatabaseName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("please provide a valid database name")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return fmt.Errorf("the database name contains illegal characters")
		}
	}
	return nil
}
