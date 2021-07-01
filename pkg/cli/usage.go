// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"flag"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

const mainUsageTemplate = `
A simple tool to collect and process web news from multiple sources.

Usage:

	{{.Name}} -config=<filename> <command> [arguments]

Main command flags:

	-config filename
		Required path to a YAML file providing the configuration parameters.
		
		"${var}" or "$var" are replaced according to the values of the current
		environment variables. References to undefined variables are replaced
		by the empty string.

The commands are:
{{range .Commands}}
	{{.Name | printf "%-19s"}} {{.Short}}{{end}}

Use "{{.Name}} help <command>" for more information about that command.
`

const commandUsageTemplate = `
Usage:

	{{.Name}} -config=<filename> {{.Command.UsageLine}}
{{.Command.Long}}
`

type mainUsageTemplateData struct {
	Name     string
	Commands []*command.Command
}

type commandUsageTemplateData struct {
	Name    string
	Command *command.Command
}

func printMainUsage(wr io.Writer, fs *flag.FlagSet) {
	t := template.New("mainUsageTemplate")
	template.Must(t.Parse(mainUsageTemplate))

	mustExecuteTemplate(t, wr, mainUsageTemplateData{
		Name:     filepath.Base(fs.Name()),
		Commands: commands,
	})
}

func printCommandUsage(wr io.Writer, fs *flag.FlagSet, cmd *command.Command) {
	t := template.New("commandUsageTemplate")
	template.Must(t.Parse(commandUsageTemplate))

	mustExecuteTemplate(t, wr, commandUsageTemplateData{
		Name:    filepath.Base(fs.Name()),
		Command: cmd,
	})
}

func mustExecuteTemplate(t *template.Template, wr io.Writer, data interface{}) {
	err := t.Execute(wr, data)
	if err != nil {
		panic(err)
	}
}

func printMainErrorAndUsage(err error, fs *flag.FlagSet) {
	_, printErr := fmt.Fprintf(os.Stderr, "%v\n", err)
	if printErr != nil {
		panic(err)
	}
	printMainUsage(os.Stderr, fs)
}

func printErrorAndCommandUsage(err error, fs *flag.FlagSet, cmd *command.Command) {
	_, printErr := fmt.Fprintf(os.Stderr, "%v\n", err)
	if printErr != nil {
		panic(err)
	}
	if command.IsInvalidArguments(err) {
		printCommandUsage(os.Stderr, fs, cmd)
	}
}
