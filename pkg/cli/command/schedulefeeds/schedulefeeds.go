// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedulefeeds

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdScheduleFeeds implements the command "whatsnew schedule-feeds".
var CmdScheduleFeeds = &command.Command{
	Name:      "schedule-feeds",
	UsageLine: "schedule-feeds",
	Short:     "periodically schedule all feeds for fetching",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew schedule-feeds".
func Run(ctx context.Context, conf *config.Config, args []string) error {
	panic("not implemented")
}
