// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package classifytext

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/textclassifier"
)

// CmdClassifyText implements the command "whatsnew classify-text".
var CmdClassifyText = &command.Command{
	Name:      "classify-text",
	UsageLine: "classify-text",
	Short:     "classify web articles",
	Long: `
The command "classify-text" runs the worker for performing text classification
of existing WebArticles.
`,
	Run: Run,
}

// Run runs the command "whatsnew classify-text".
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

	classifierConn, err := grpcconn.Dial(ctx, conf.Workers.TextClassifier.ClassifierServer)
	if err != nil {
		return err
	}

	fk, err := workers.NewManager(conf.Faktory)
	if err != nil {
		return err
	}

	zsc := textclassifier.New(conf.Workers.TextClassifier, db, classifierConn, fk)
	zsc.Run()

	return nil
}
