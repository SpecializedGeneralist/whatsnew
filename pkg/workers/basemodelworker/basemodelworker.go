// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package basemodelworker

import (
	"context"
	"fmt"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Worker can be embed by specific worker implementations which
// operate on database models.
type Worker struct {
	Name        string
	DB          *gorm.DB
	FK          *faktory_worker.Manager
	Log         zerolog.Logger
	Concurrency int
	Queues      []string
	Perform     Perform
}

// Perform actually executes the job.
type Perform func(ctx context.Context, modelID uint) error

// Run sets the concurrency, registers the worker handler and starts
// processing jobs.
// This function never returns (refer to faktory_worker_go Manager.Run).
func (w Worker) Run() {
	w.FK.Concurrency = w.Concurrency
	w.FK.Register(w.Name, w.faktoryPerform)
	w.FK.ProcessStrictPriorityQueues(w.Queues...)
	w.FK.Labels = []string{w.Name}
	w.FK.Run()
}

func (w Worker) faktoryPerform(ctx context.Context, args ...interface{}) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid arguments: %#v", args)
	}

	f, ok := args[0].(float64)
	if !ok {
		return fmt.Errorf("invalid model ID argument: %#v", args[0])
	}
	modelID := uint(f)

	return w.Perform(ctx, modelID)
}
