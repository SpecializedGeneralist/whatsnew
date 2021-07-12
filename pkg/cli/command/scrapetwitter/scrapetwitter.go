// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scrapetwitter

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/twitterscraper"
)

// CmdScrapeTwitter implements the command "whatsnew scrape-twitter".
var CmdScrapeTwitter = &command.Command{
	Name:      "scrape-twitter",
	UsageLine: "scrape-twitter",
	Short:     "scrape twitter from user or search sources",
	Long: `
The command "scrape-twitter" runs the worker for fetching tweets from specific
twitter sources.
`,
	Run: Run,
}

// Run runs the command "whatsnew scrape-twitter".
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

	ts := twitterscraper.New(conf.Workers.TwitterScraper, db, fk)
	ts.Run()

	return nil
}
