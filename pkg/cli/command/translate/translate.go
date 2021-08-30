// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package translate

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/translator"
)

// CmdTranslate implements the command "whatsnew translate".
var CmdTranslate = &command.Command{
	Name:      "translate",
	UsageLine: "translate",
	Short:     "translate web articles with SpecializedGeneralist translation service",
	Long: `
The command "translate" runs the worker for performing translating the title
of existing WebArticles.
`,
	Run: Run,
}

// Run runs the command "whatsnew translate".
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

	translatorConn, err := grpcconn.Dial(ctx, conf.Workers.Translator.TranslatorServer)
	if err != nil {
		return err
	}

	fk, err := workers.NewManager(conf.Faktory)
	if err != nil {
		return err
	}

	zsc := translator.New(conf.Workers.Translator, db, translatorConn, fk)
	zsc.Run()

	return nil
}
