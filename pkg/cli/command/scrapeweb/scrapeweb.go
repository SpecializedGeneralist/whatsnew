// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scrapeweb

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdScrapeWeb implements the command "whatsnew scrape-web".
var CmdScrapeWeb = &command.Command{
	Name:      "scrape-web",
	UsageLine: "scrape-web",
	Short:     "scrape news articles from Web Resource URLs",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew scrape-web".
func Run(ctx context.Context, conf *config.Config, args []string) error {
	panic("not implemented")
}
