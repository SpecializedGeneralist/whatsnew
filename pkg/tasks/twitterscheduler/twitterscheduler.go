// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package twitterscheduler

import (
	"context"
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

// TwitterScheduler implements the mechanism for periodically fetching all
// enabled TwitterSources and scheduling a set of jobs for each TwitterSource.
type TwitterScheduler struct {
	conf config.TwitterScheduler
	db   *gorm.DB
	fk   *faktory.Client
	log  zerolog.Logger
}

// New creates a new TwitterScheduler.
func New(conf config.TwitterScheduler, db *gorm.DB, fk *faktory.Client) *TwitterScheduler {
	return &TwitterScheduler{
		conf: conf,
		db:   db,
		fk:   fk,
		log:  log.Logger.Level(zerolog.Level(conf.LogLevel)),
	}
}

const batchSize = 100

var errStop = errors.New("stop")

// Run starts the twitter-sources scheduling process.
//
// This function should ideally run forever, unless an error is encountered
// or the context is done.
func (fs *TwitterScheduler) Run(ctx context.Context) (err error) {
	fs.log.Info().Msg("twitter-sources scheduling starts")

Loop:
	for {
		err = fs.findAndScheduleSources(ctx)
		if err != nil {
			break
		}

		fs.log.Info().Msgf("waiting %s", fs.conf.TimeInterval)
		select {
		case <-time.After(fs.conf.TimeInterval):
		case <-ctx.Done():
			fs.log.Warn().Msg("context done")
			break Loop
		}
	}

	if err != nil && err != errStop {
		fs.log.Err(err).Msg("twitter-sources scheduling ends with error")
		return err
	}

	fs.log.Info().Msg("twitter-sources scheduling ends")
	return nil
}

func (fs *TwitterScheduler) findAndScheduleSources(ctx context.Context) error {
	fs.log.Info().Msg("scheduling all twitter-sources")

	query := fs.db.WithContext(ctx).
		Where("enabled = true").
		Order("last_retrieved_at NULLS FIRST, id")

	var sources []*models.TwitterSource
	res := query.FindInBatches(&sources, batchSize, func(_ *gorm.DB, batch int) error {
		fs.log.Debug().Msgf("batch %d", batch)
		return fs.processBatch(ctx, sources)
	})
	return res.Error
}

func (fs *TwitterScheduler) processBatch(ctx context.Context, sources []*models.TwitterSource) error {
	for _, source := range sources {
		if ctxIsDone(ctx) {
			fs.log.Warn().Msg("context done")
			return errStop
		}
		err := fs.scheduleSourceJobs(source)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *TwitterScheduler) scheduleSourceJobs(source *models.TwitterSource) error {
	// The Context is ignored on purpose here, so that it is more likely that
	// the full set of jobs is scheduled for each twitter-source, even if the
	// context is canceled in the meanwhile.

	for _, fj := range fs.conf.Jobs {
		job := faktory.NewJob(fj.JobType, source.ID)
		job.Queue = fj.Queue
		job.ReserveFor = fj.ReserveFor
		job.Retry = new(int)
		*job.Retry = fj.Retry

		fs.log.Trace().Interface("job", job).Msg("push job")

		err := fs.fk.Push(job)
		if err != nil {
			return fmt.Errorf("error pushing Job %+v for TwitterSource %d: %w", fj, source.ID, err)
		}
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
