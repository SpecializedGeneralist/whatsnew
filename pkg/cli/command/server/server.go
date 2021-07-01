// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdServer implements the command "whatsnew server".
var CmdServer = &command.Command{
	Name:      "server",
	UsageLine: "server",
	Short:     "run the API server",
	Long:      ``, // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew server".
func Run(conf *config.Config, args []string) error {
	panic("not implemented")
}
