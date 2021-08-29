// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scheduletwitter

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/twitterscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
)

// CmdScheduleTwitter implements the command "whatsnew schedule-twitter".
var CmdScheduleTwitter = &command.Command{
	Name:      "schedule-twitter",
	UsageLine: "schedule-twitter",
	Short:     "periodically schedule all Twitter sources for scraping",
	Long: `
The command "schedule-twitter" starts a process which periodically fetches
all enabled TwitterSources from the database and schedules new jobs for each 
of them.
`,
	Run: Run,
}

// Run runs the command "whatsnew schedule-twitter".
func Run(ctx context.Context, conf *config.Config, args []string) (err error) {
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

	fs := twitterscheduler.New(conf.TwitterScheduler, db, fk)
	return fs.Run(ctx)
}
