// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/feedsimporting"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) addFeeds() *cli.Command {
	return &cli.Command{
		Name:  "add-feeds",
		Usage: "Add new feeds from a list",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Required: true,
				Usage:    "load list of feed URLs from `FILE`",
			},
		},
		Action: func(c *cli.Context) error {
			db, err := models.OpenDB(app.config.DB.DSN, app.newContextLogger(c))
			if err != nil {
				return err
			}
			return feedsimporting.AddFeedsFromFile(db, c.String("file"))
		},
	}
}
