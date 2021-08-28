// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package detectduplicates

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/hnswclient"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/duplicatedetector"
)

// CmdDetectDuplicates implements the command "whatsnew detect-duplicates".
var CmdDetectDuplicates = &command.Command{
	Name:      "detect-duplicates",
	UsageLine: "detect-duplicates",
	Short:     "perform near-duplicate news detection via cosine similarity",
	Long: `
The command "detect-duplicates" runs the worker for performing near-duplicate
detection over existing WebArticles.
`,
	Run: Run,
}

// Run runs the command "whatsnew detect-duplicates".
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

	hnswConn, err := grpcconn.Dial(ctx, conf.HNSW.Server)
	if err != nil {
		return err
	}

	hnswClient := hnswclient.New(hnswConn, conf.HNSW.Index)

	fk, err := workers.NewManager(conf.Faktory)
	if err != nil {
		return err
	}

	dd := duplicatedetector.New(conf.Workers.DuplicateDetector, db, hnswClient, fk)
	dd.Run()

	return nil
}
