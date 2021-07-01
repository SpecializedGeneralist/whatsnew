// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zeroshotclassify

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdZeroShotClassify implements the command "whatsnew zero-shot-classify".
var CmdZeroShotClassify = &command.Command{
	Name:      "zero-shot-classify",
	UsageLine: "zero-shot-classify",
	Short:     "classify articles with spaGO zero-shot classification service",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew zero-shot-classify".
func Run(conf *config.Config, args []string) error {
	panic("not implemented")
}
