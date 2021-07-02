// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vectorize

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// CmdVectorize implements the command "whatsnew vectorize".
var CmdVectorize = &command.Command{
	Name:      "vectorize",
	UsageLine: "vectorize",
	Short:     "vectorize web articles through LaBSE encoding",
	Long:      "", // TODO: ...
	Run:       Run,
}

// Run runs the command "whatsnew vectorize".
func Run(ctx context.Context, conf *config.Config, args []string) error {
	panic("not implemented")
}
