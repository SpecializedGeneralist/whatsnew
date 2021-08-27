// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vectorize

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/hnswclient"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/vectorizer"
)

// CmdVectorize implements the command "whatsnew vectorize".
var CmdVectorize = &command.Command{
	Name:      "vectorize",
	UsageLine: "vectorize",
	Short:     "vectorize web articles through LaBSE encoding",
	Long: `
The command "vectorize" runs the worker for creating a vector representation
of existing WebArticles, storing the result on the configured HNSW server.
`,
	Run: Run,
}

// Run runs the command "whatsnew vectorize".
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

	bertConn, err := grpcconn.Dial(ctx, conf.Workers.Vectorizer.SpagoBERTServer)
	if err != nil {
		return err
	}

	hnswConn, err := grpcconn.Dial(ctx, conf.HNSW.Server)
	if err != nil {
		return err
	}

	hnswClient := hnswclient.New(hnswConn, conf.HNSW.Index)

	fk, err := workers.NewManager(conf.Faktory)
	if err != nil {
		return err
	}

	v := vectorizer.New(conf.Workers.Vectorizer, db, bertConn, hnswClient, fk)
	v.Run()

	return nil
}
