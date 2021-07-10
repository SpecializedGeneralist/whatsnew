// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fetchgdelt

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/gdeltfetcher"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
)

// CmdFetchGDELT implements the command "whatsnew fetch-gdelt".
var CmdFetchGDELT = &command.Command{
	Name:      "fetch-gdelt",
	UsageLine: "fetch-gdelt",
	Short:     "fetch latest news from GDELT",
	Long: `
The command "fetch-gdelt" starts a process which periodically fetches
events from GDELT Master CSV Data File List. For each event, the source
URL is extracted (a link to the first news report it found this event in).
If present, a new WebResource is created and new related jobs are scheduled 
(as configured).
`,
	Run: Run,
}

// Run runs the command "whatsnew fetch-gdelt".
func Run(ctx context.Context, conf *config.Config, args []string) error {
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

	fk, err := workers.NewClient(conf.Faktory)
	if err != nil {
		return err
	}
	defer func() {
		if e := fk.Close(); e != nil && err == nil {
			err = e
		}
	}()

	fs := gdeltfetcher.New(conf.GDELTFetcher, db, fk)
	return fs.Run(ctx)
}
