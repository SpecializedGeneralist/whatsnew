// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scrapetwitter

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdScrapeTwitter implements the command "whatsnew scrape-twitter".
var CmdScrapeTwitter = &command.Command{
	Name:      "scrape-twitter",
	UsageLine: "scrape-twitter",
	Short:     "scrape twitter from user or search sources",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew scrape-twitter".
func Run(ctx context.Context, conf *config.Config, args []string) error {
	panic("not implemented")
}
