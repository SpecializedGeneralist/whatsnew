// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"github.com/nlpodyssey/whatsnew/pkg/models"
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
			dbName := c.String("name")
			result := db.Exec("CREATE DATABASE ?;", dbName)
			if result.Error != nil {
				fmt.Printf("Unable to create `%s`, it may already exists...", dbName)
				return nil
			}
			fmt.Printf("Done!")
			return nil
		},
	}
}
