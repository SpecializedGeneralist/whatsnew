// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package purgehnsw

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/hnswpurger"
)

// CmdPurgeHNSW implements the command "purge-hnsw".
var CmdPurgeHNSW = &command.Command{
	Name:      "purge-hnsw",
	UsageLine: "purge-hnsw",
	Short:     "delete old HNSW indices",
	Long: `
The command "purge-hnsw" runs a process that periodically removes old
indices from the HNSW server.
`,
	Run: Run,
}

// Run runs the command "purge-hnsw".
func Run(ctx context.Context, conf *config.Config, args []string) error {
	if len(args) != 0 {
		return command.ErrInvalidArguments
	}

	hp := hnswpurger.New(conf.Tasks.HNSWPurger, conf.HNSW)
	return hp.Run(ctx)
}
