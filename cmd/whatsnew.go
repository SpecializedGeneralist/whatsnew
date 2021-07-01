// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli"
	"os"
)

func main() {
	err := cli.Run(os.Args[0], os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
