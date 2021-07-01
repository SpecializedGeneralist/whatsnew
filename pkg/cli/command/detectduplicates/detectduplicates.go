// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package detectduplicates

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdDetectDuplicates implements the command "whatsnew detectd-uplicates".
var CmdDetectDuplicates = &command.Command{
	Name:      "detectd-uplicates",
	UsageLine: "detectd-uplicates",
	Short:     "perform near-duplicate news detection via cosine similarity",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew detectd-uplicates".
func Run(conf *config.Config, args []string) error {
	panic("not implemented")
}
