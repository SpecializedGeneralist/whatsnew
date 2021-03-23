// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/twittersourcesimporting"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) addTwitterSources() *cli.Command {
	return &cli.Command{
		Name:  "add-twitter-sources",
		Usage: "Add new Twitter sources from a TSV file (columns: [type, value])",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Required: true,
				Usage:    "load Twitter sources from TSV `FILE`",
			},
		},
		Action: func(c *cli.Context) error {
			db, err := models.OpenDB(app.config.DB.DSN, app.newContextLogger(c))
			if err != nil {
				return err
			}
			return twittersourcesimporting.AddTwitterSourcesFromTSVFile(db, c.String("file"))
		},
	}
}
