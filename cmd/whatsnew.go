// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func main() {
	initLogger()
	err := cli.Run(os.Args[0], os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}

// Initialize the global zerolog Logger, which will be used as basis by all
// whatsnew commands and operations.
func initLogger() {
	w := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	log.Logger = log.Output(w)
}
