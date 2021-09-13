// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"context"
	"errors"
	"flag"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/classifytext"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/db"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/detectduplicates"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/extractinformation"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/fetchfeeds"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/fetchgdelt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/parsegeo"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/recoverjobs"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/schedulefeeds"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/scheduletwitter"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/scrapetwitter"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/scrapeweb"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/server"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/translate"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/vectorize"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command/zeroshotclassify"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"os"
)

var (
	errMissingCommand = errors.New("a command is missing")
	errInvalidCommand = errors.New("the command is invalid")
	errMissingConfig  = errors.New("the configuration file is missing")

	// commands is the list of all whatsnew Commands.
	commands = []*command.Command{
		db.CmdDB,
		server.CmdServer,
		schedulefeeds.CmdScheduleFeeds,
		scheduletwitter.CmdScheduleTwitter,
		fetchfeeds.CmdFetchFeeds,
		fetchgdelt.CmdFetchGDELT,
		scrapetwitter.CmdScrapeTwitter,
		scrapeweb.CmdScrapeWeb,
		translate.CmdTranslate,
		zeroshotclassify.CmdZeroShotClassify,
		classifytext.CmdClassifyText,
		parsegeo.CmdParseGeo,
		vectorize.CmdVectorize,
		detectduplicates.CmdDetectDuplicates,
		extractinformation.CmdExtractInformation,
		recoverjobs.CmdRecoverJobs,
	}
)

// Run is the main entry point for the Command Line Interface.
//
// It accepts the name of the program, which is used for help messages, and a
// list of arguments.
func Run(ctx context.Context, name string, arguments []string) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = func() { printMainUsage(os.Stderr, fs) }
	configFilename := fs.String("config", "", "")

	err := fs.Parse(arguments)
	if err == flag.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}

	args := fs.Args()
	if len(args) == 0 {
		printMainErrorAndUsage(errMissingCommand, fs)
		return errMissingCommand
	}
	cmdName, cmdArgs := args[0], args[1:]

	if cmdName == "help" {
		return runHelpCommand(fs, cmdArgs)
	}

	conf, err := loadConfig(*configFilename)
	if err != nil {
		printMainErrorAndUsage(err, fs)
		return err
	}

	for _, cmd := range commands {
		if cmd.Name != cmdName {
			continue
		}
		err = cmd.Run(ctx, conf, cmdArgs)
		if err != nil {
			printErrorAndCommandUsage(err, fs, cmd)
		}
		return err
	}

	printMainErrorAndUsage(errInvalidCommand, fs)
	return errInvalidCommand
}

func loadConfig(filename string) (*config.Config, error) {
	if filename == "" {
		return nil, errMissingConfig
	}
	conf, err := config.FromYAMLFile(filename)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
