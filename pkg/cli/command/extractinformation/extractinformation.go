// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package extractinformation

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/informationextractor"
)

// CmdExtractInformation implements the command "whatsnew extract-information".
var CmdExtractInformation = &command.Command{
	Name:      "extract-information",
	UsageLine: "extract-information",
	Short:     "extract specific information from web articles with spaGO BERT Question Answering",
	Long: `
The command "extract-information" runs the worker for attempting information
extraction over existing WebArticles, making use of spaGO BERT Question
Answering service.
`,
	Run: Run,
}

// Run runs the command "whatsnew extract-information".
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

	ie := informationextractor.New(conf.Workers.InformationExtractor, db, fk)
	ie.Run()

	return nil
}
