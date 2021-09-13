// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parsegeo

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/geoparser"
)

// CmdParseGeo implements the command "whatsnew parse-geo".
var CmdParseGeo = &command.Command{
	Name:      "parse-geo",
	UsageLine: "parse-geo",
	Short:     "classify web articles",
	Long: `
The command "parse-geo" runs the worker for extracting geo-political entities
from existing WebArticles.
`,
	Run: Run,
}

// Run runs the command "whatsnew parse-geo".
func Run(_ context.Context, conf *config.Config, args []string) error {
	if len(args) != 0 {
		return command.ErrInvalidArguments
	}

	db, err := database.OpenDB(conf.DB)
	if err != nil {
		return err
	}
	defer func() {
		if e := database.CloseDB(db); e != nil && err == nil {
			err = e
		}
	}()

	fk, err := workers.NewManager(conf.Faktory)
	if err != nil {
		return err
	}

	gp := geoparser.New(conf.Workers.GeoParser, db, fk)
	gp.Run()

	return nil
}
