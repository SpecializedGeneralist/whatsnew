// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package workers

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	faktory "github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"os"
)

// NewManager instantiates and prepares a new faktory Manager.
//
// It also sets the FAKTORY_URL variable globally: for this reason, you cannot
// create multiple faktory managers or clients pointing to different servers
// from the same instance of the program.
func NewManager(conf config.Faktory) (*faktory_worker.Manager, error) {
	err := setFaktoryURL(conf.URL)
	if err != nil {
		return nil, err
	}
	mgr := faktory_worker.NewManager()
	mgr.ProcessStrictPriorityQueues("default")
	mgr.Logger = NewFaktoryLogger(zerolog.Level(conf.LogLevel))
	return mgr, nil
}

// NewClient returns a new faktory Client.
//
// It sets the FAKTORY_URL variable globally: for this reason, you cannot
// create multiple faktory managers or clients pointing to different servers
// from the same instance of the program.
func NewClient(conf config.Faktory) (*faktory.Client, error) {
	err := setFaktoryURL(conf.URL)
	if err != nil {
		return nil, err
	}
	cl, err := faktory.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening faktory client: %w", err)
	}
	return cl, nil
}

func setFaktoryURL(url string) error {
	err := os.Setenv("FAKTORY_URL", url)
	if err != nil {
		return fmt.Errorf("error setting FAKTORY_URL variable: %w", err)
	}
	return nil
}
