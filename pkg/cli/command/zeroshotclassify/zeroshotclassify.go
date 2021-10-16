// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zeroshotclassify

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/zeroshotclassifier"
)

// CmdZeroShotClassify implements the command "whatsnew zero-shot-classify".
var CmdZeroShotClassify = &command.Command{
	Name:      "zero-shot-classify",
	UsageLine: "zero-shot-classify",
	Short:     "classify articles with spaGO BART zero-shot classification service",
	Long: `
The command "zero-shot-classify" runs the worker for performing BART zero-shot
classification of existing WebArticles.
`,
	Run: Run,
}

// Run runs the command "whatsnew zero-shot-classify".
func Run(_ context.Context, conf *config.Config, args []string) error {
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

	fk, err := workers.NewManager(conf.Faktory)
	if err != nil {
		return err
	}

	zsc := zeroshotclassifier.New(conf.Workers.ZeroShotClassifier, db, fk)
	zsc.Run()

	return nil
}
