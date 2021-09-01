// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server"
)

// CmdServer implements the command "whatsnew server".
var CmdServer = &command.Command{
	Name:      "server",
	UsageLine: "server",
	Short:     "run the API server",
	Long: `
The command "server" runs the HTTP + gRPC server for performing basic operations
on the database models.
`,
	Run: Run,
}

// Run runs the command "whatsnew server".
func Run(ctx context.Context, conf *config.Config, args []string) error {
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

	s := server.New(conf.Server, db)
	return s.Run(ctx)
}
