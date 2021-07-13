// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scrapeweb

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/webscraper"
)

// CmdScrapeWeb implements the command "whatsnew scrape-web".
var CmdScrapeWeb = &command.Command{
	Name:      "scrape-web",
	UsageLine: "scrape-web",
	Short:     "scrape news articles from Web Resource URLs",
	Long: `
The command "scrape-web" runs the worker for scraping Web pages, creating new
WebArticles from existing WebResources.
`,
	Run: Run,
}

// Run runs the command "whatsnew scrape-web".
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

	ws := webscraper.New(conf.Workers.WebScraper, db, fk)
	ws.Run()

	return nil
}
