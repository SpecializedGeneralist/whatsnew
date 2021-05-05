// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) runServer() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Run API server",
		Action: func(c *cli.Context) error {
			logger := app.newContextLogger(c)
			db, err := models.OpenDB(app.config.DB.DSN, logger)
			if err != nil {
				return err
			}
			srv := server.New(app.config.Server, db, logger)
			return srv.Run()
		},
	}
}
