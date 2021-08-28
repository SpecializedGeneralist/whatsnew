// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jobsrecoverer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	faktory "github.com/contribsys/faktory/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"time"
)

// JobsRecoverer implements the mechanism for periodically checking for old
// PendingJobs and attempting their rescheduling.
type JobsRecoverer struct {
	conf config.JobsRecoverer
	db   *gorm.DB
	fk   *faktory.Client
	log  zerolog.Logger
}

// New creates a new JobsRecoverer.
func New(conf config.JobsRecoverer, db *gorm.DB, fk *faktory.Client) *JobsRecoverer {
	return &JobsRecoverer{
		conf: conf,
		db:   db,
		fk:   fk,
		log:  log.Logger.Level(zerolog.Level(conf.LogLevel)),
	}
}

const batchSize = 100

var errStop = errors.New("stop")

// Run starts the pending jobs recovery process.
//
// This function should ideally run forever, unless an error is encountered
// or the context is done.
func (jr *JobsRecoverer) Run(ctx context.Context) (err error) {
	jr.log.Info().Msg("pending jobs recovery starts")

Loop:
	for {
		err = jr.findAndRecoverPendingJobs(ctx)
		if err != nil {
			break
		}

		jr.log.Info().Msgf("waiting %s", jr.conf.TimeInterval)
		select {
		case <-time.After(jr.conf.TimeInterval):
		case <-ctx.Done():
			jr.log.Warn().Msg("context done")
			break Loop
		}
	}

	if err != nil && err != errStop {
		jr.log.Err(err).Msg("pending jobs recovery ends with error")
		return err
	}

	jr.log.Info().Msg("pending jobs recovery ends")
	return nil
}

func (jr *JobsRecoverer) findAndRecoverPendingJobs(ctx context.Context) error {
	upperTime := time.Now().UTC().Add(-jr.conf.LeewayTime)
	query := jr.db.WithContext(ctx).
		Where("created_at < ?", upperTime).
		Order("created_at")

	var pjs []*models.PendingJob
	res := query.FindInBatches(&pjs, batchSize, func(_ *gorm.DB, batch int) error {
		jr.log.Trace().Msgf("batch %d", batch)
		return jr.processBatch(ctx, pjs)
	})
	return res.Error
}

func (jr *JobsRecoverer) processBatch(ctx context.Context, pjs []*models.PendingJob) error {
	for _, pj := range pjs {
		if ctxIsDone(ctx) {
			jr.log.Warn().Msg("context done")
			return errStop
		}
		err := jr.recoverJob(pj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (jr *JobsRecoverer) recoverJob(pj *models.PendingJob) error {
	logger := jr.log.With().Str("JID", pj.ID).Logger()
	logger.Debug().Msg("recovering job")

	var job *faktory.Job
	err := json.Unmarshal(pj.Data, &job)
	if err != nil {
		logger.Err(err).Msg("json.Unmarshal error")
		return nil
	}

	err = jr.fk.Push(job)
	if err != nil {
		return fmt.Errorf("error pushing job %s: %w", job.Jid, err)
	}

	res := jr.db.Delete(pj)
	if res.Error != nil {
		return fmt.Errorf("error deleting PendingJob %s: %w", pj.ID, err)
	}
	return nil
}

func ctxIsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
