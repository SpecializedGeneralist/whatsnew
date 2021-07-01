// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fetchtweets

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdFetchTweets implements the command "whatsnew fetch-tweets".
var CmdFetchTweets = &command.Command{
	Name:      "fetch-tweets",
	UsageLine: "fetch-tweets",
	Short:     "fetch tweets by username or search query",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "".
func Run(conf *config.Config, args []string) error {
	panic("not implemented")
}
