// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feedscheduler

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

// FeedScheduler implements the mechanism for periodically fetching all
// enabled Feeds and scheduling a set of jobs for each Feed.
type FeedScheduler struct {
	conf config.FeedScheduler
	db   *gorm.DB
	fk   *faktory.Client
	log  zerolog.Logger
}

// New creates a new FeedScheduler.
func New(conf config.FeedScheduler, db *gorm.DB, fk *faktory.Client) *FeedScheduler {
	return &FeedScheduler{
		conf: conf,
		db:   db,
		fk:   fk,
		log:  log.Logger.Level(zerolog.Level(conf.LogLevel)),
	}
}

const batchSize = 100

var errStop = errors.New("stop")

// Run starts the feed scheduling process.
//
// This function should ideally run forever, unless an error is encountered
// or the context is done.
func (fs *FeedScheduler) Run(ctx context.Context) (err error) {
	fs.log.Info().Msg("feed scheduling starts")

Loop:
	for {
		err = fs.findAndScheduleFeeds(ctx)
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
		fs.log.Err(err).Msg("feed scheduling ends with error")
		return err
	}

	fs.log.Info().Msg("feed scheduling ends")
	return nil
}

func (fs *FeedScheduler) findAndScheduleFeeds(ctx context.Context) error {
	fs.log.Info().Msg("scheduling all feeds")

	query := fs.db.WithContext(ctx).
		Where("enabled = true").
		Order("last_retrieved_at, id")

	var feeds []*models.Feed
	res := query.FindInBatches(&feeds, batchSize, func(_ *gorm.DB, batch int) error {
		fs.log.Debug().Msgf("batch %d", batch)
		return fs.processBatch(ctx, feeds)
	})
	return res.Error
}

func (fs *FeedScheduler) processBatch(ctx context.Context, feeds []*models.Feed) error {
	for _, feed := range feeds {
		if ctxIsDone(ctx) {
			fs.log.Warn().Msg("context done")
			return errStop
		}
		err := fs.scheduleFeedJobs(feed)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FeedScheduler) scheduleFeedJobs(feed *models.Feed) error {
	// The Context is ignored on purpose here, so that it is more likely that
	// the full set of jobs is scheduled for each feed, even if the context
	// is canceled in the meanwhile.

	for _, fj := range fs.conf.Jobs {
		job := faktory.NewJob(fj.JobType, feed.ID)
		job.Queue = fj.Queue
		job.ReserveFor = fj.ReserveFor
		job.Retry = fj.Retry

		fs.log.Trace().Interface("job", job).Msg("push job")

		err := fs.fk.Push(job)
		if err != nil {
			return fmt.Errorf("error pushing Job %+v for feed %d: %w", fj, feed.ID, err)
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
