// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fetchfeeds

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/feedfetcher"
)

// CmdFetchFeeds implements the command "whatsnew fetch-feeds".
var CmdFetchFeeds = &command.Command{
	Name:      "fetch-feeds",
	UsageLine: "fetch-feeds",
	Short:     "run the worker to fetch feeds and get new feed items",
	Long: `
The "fetch-feeds" command runs the worker for fetching feeds and getting
new feed items.
`,
	Run: Run,
}

// Run runs the command "whatsnew fetch-feeds".
func Run(_ context.Context, conf *config.Config, args []string) (err error) {
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

	ff := feedfetcher.New(conf.Workers.FeedFetcher, db, fk)
	ff.Run()

	return nil
}
