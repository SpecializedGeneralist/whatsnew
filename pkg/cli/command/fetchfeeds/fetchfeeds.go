// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fetchfeeds

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdFetchFeeds implements the command "whatsnew fetch-feeds".
var CmdFetchFeeds = &command.Command{
	Name:      "fetch-feeds",
	UsageLine: "fetch-feeds",
	Short:     "fetch all feeds and get new feed items",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew fetch-feeds".
func Run(ctx context.Context, conf *config.Config, args []string) error {
	panic("not implemented")
}
