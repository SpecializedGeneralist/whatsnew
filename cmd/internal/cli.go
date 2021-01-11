// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"
)

// CLIApp allows to run WhatsNew's tasks from command line.
type CLIApp struct {
	cliApp *cli.App
	config configuration.Configuration
	logger *log.Logger
}

// NewCLIApp creates a new CLIApp.
func NewCLIApp() *CLIApp {
	app := &CLIApp{
		logger: log.New(os.Stderr, "", 0),
	}
	app.initCLIApp()
	return app
}

// Run is the entry point for WhatsNew command line interface.
func (app *CLIApp) Run() {
	err := app.cliApp.Run(os.Args)
	if err != nil {
		app.logger.Fatalf("Application error: %v", err)
	}
}

func (app *CLIApp) initCLIApp() {
	app.cliApp = &cli.App{
		Name:                 "whatsnew",
		HelpName:             "whatsnew",
		Usage:                "A simple tool to collect and process quite a few web news from multiple sources",
		UsageText:            "",
		BashComplete:         cli.DefaultAppComplete,
		EnableBashCompletion: true,
		Action:               defaultAction,
		Before: func(c *cli.Context) error {
			config, err := configuration.FromYAMLFile(c.String("config"))
			if err != nil {
				return err
			}
			app.config = config
			return nil
		},
		Flags:    globalFlags,
		Commands: app.commands(),
	}
}

func defaultAction(c *cli.Context) error {
	err := cli.ShowAppHelp(c)
	if err != nil {
		return err
	}

	message := "missing command"
	if c.Args().Present() {
		message = "invalid command"
	}
	return cli.Exit(message, 1)
}

var globalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "config",
		Aliases:  []string{"c"},
		Required: true,
		Usage:    "load configuration from YAML `FILE`",
	},
}

func (app *CLIApp) newLogger() zerolog.Logger {
	w := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	return zerolog.New(w).With().Timestamp().Logger().Level(zerolog.Level(app.config.LogLevel))
}

func (app *CLIApp) newContextLogger(c *cli.Context) zerolog.Logger {
	l := app.newLogger()
	if c.Command != nil && len(c.Command.Name) > 0 {
		l = l.With().Str("command", c.Command.Name).Logger()
	}
	return l
}

func (app *CLIApp) commands() []*cli.Command {
	return []*cli.Command{
		app.createDB(),
		app.migrateDB(),
		app.addFeeds(),
		app.fetchFeeds(),
		app.fetchGDELT(),
		app.scrapeWeb(),
		app.duplicateDetector(),
	}
}
