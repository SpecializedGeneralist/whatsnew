// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"errors"
	"flag"
	"os"
)

var (
	errInvalidHelpArgs    = errors.New("invalid help arguments")
	errUnknownHelpCommand = errors.New("invalid help command or arguments")
)

func runHelpCommand(fs *flag.FlagSet, args []string) error {
	if len(args) == 0 {
		printMainUsage(os.Stdout, fs)
		return nil
	}
	if len(args) > 1 {
		printMainErrorAndUsage(errInvalidHelpArgs, fs)
		return errInvalidHelpArgs
	}

	arg := args[0]
	for _, cmd := range commands {
		if cmd.Name != arg {
			continue
		}
		printCommandUsage(os.Stdout, fs, cmd)
		return nil
	}

	printMainErrorAndUsage(errUnknownHelpCommand, fs)
	return errUnknownHelpCommand
}
