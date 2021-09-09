// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package recoverjobs

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/jobsrecoverer"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
)

// CmdRecoverJobs implements the command "whatsnew recover-jobs".
var CmdRecoverJobs = &command.Command{
	Name:      "recover-jobs",
	UsageLine: "recover-jobs",
	Short:     "periodically check for pending jobs and reschedule them",
	Long: `
The command "recover-jobs" starts a process which periodically attempts
the recovery (re-scheduling) of pending jobs.
`,
	Run: Run,
}

// Run runs the command "whatsnew recover-jobs".
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

	jr := jobsrecoverer.New(conf.Tasks.JobsRecoverer, db, fk)
	return jr.Run(ctx)
}
