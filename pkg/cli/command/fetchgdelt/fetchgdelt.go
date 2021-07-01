// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fetchgdelt

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdFetchGDELT implements the command "whatsnew fetch-gdelt".
var CmdFetchGDELT = &command.Command{
	Name:      "fetch-gdelt",
	UsageLine: "fetch-gdelt",
	Short:     "fetch latest news from GDELT",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew fetch-gdelt".
func Run(conf *config.Config, args []string) error {
	panic("not implemented")
}
