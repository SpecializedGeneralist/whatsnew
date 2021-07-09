// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"encoding/json"
	"fmt"
	faktory "github.com/contribsys/faktory/client"
	"gorm.io/datatypes"
	"time"
)

// A PendingJob represents a Faktory job which is not yet guaranteed to be
// scheduled (i.e. pushed to the server).
//
// Please refer to jobscheduler package documentation to understand the
// benefits and use cases of this model.
type PendingJob struct {
	// ID corresponds to the Job ID.
	ID string `gorm:"not null;primaryKey"`

	// The creation time can be useful for implementing a recovery process,
	// which could look for the existence of PendingJobs older than
	// a certain leeway timespan.
	CreatedAt time.Time `gorm:"not null;index"`

	// JSON-encoded Faktory Job.
	Data datatypes.JSON `gorm:"not null"`
}

// NewPendingJob builds a new PendingJob, setting ID to the job.Jid and
// Data to the JSON serialization of the job.
//
// It returns an error if json.Marshal fails.
func NewPendingJob(job *faktory.Job) (*PendingJob, error) {
	data, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("error marshaling job: %w", err)
	}
	return &PendingJob{ID: job.Jid, Data: data}, nil
}
