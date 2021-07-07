// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feedscheduling

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

// FeedsScheduling implements the mechanism for periodically fetching all
// enabled Feeds and scheduling a set of jobs for each Feed.
type FeedsScheduling struct {
	conf config.FeedsScheduling
	db   *gorm.DB
	fk   *faktory.Client
	log  zerolog.Logger
}

// New creates a new FeedsScheduling.
func New(conf config.FeedsScheduling, db *gorm.DB, fk *faktory.Client) *FeedsScheduling {
	return &FeedsScheduling{
		conf: conf,
		db:   db,
		fk:   fk,
		log:  log.Logger.Level(zerolog.Level(conf.LogLevel)),
	}
}

const batchSize = 100

var errStop = errors.New("stop")

// Run starts the feeds scheduling process.
//
// This function should ideally run forever, unless an error is encountered
// or the context is done.
func (fs *FeedsScheduling) Run(ctx context.Context) (err error) {
	fs.log.Info().Msg("feeds scheduling starts")

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
		fs.log.Err(err).Msg("feeds scheduling ends with error")
		return err
	}

	fs.log.Info().Msg("feeds scheduling ends")
	return nil
}

func (fs *FeedsScheduling) findAndScheduleFeeds(ctx context.Context) error {
	fs.log.Info().Msg("scheduling all feeds")

	query := fs.db.WithContext(ctx).
		Where("enabled = true").
		Order("last_retrieved_at, id")

	var feeds []*models.Feed
	res := query.FindInBatches(&feeds, batchSize, func(_ *gorm.DB, batch int) error {
		fs.log.Trace().Msgf("batch %d", batch)
		return fs.processBatch(ctx, feeds)
	})
	return res.Error
}

func (fs *FeedsScheduling) processBatch(ctx context.Context, feeds []*models.Feed) error {
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

func (fs *FeedsScheduling) scheduleFeedJobs(feed *models.Feed) error {
	// The Context is ignored on purpose here, so that it is more likely that
	// the full set of jobs is scheduled for each feed, even if the context
	// is canceled in the meanwhile.

	for _, jobType := range fs.conf.Jobs {
		job := faktory.NewJob(jobType, feed.ID)
		job.Retry = 0 // No retries, since it will be called periodically

		fs.log.Trace().Interface("job", job).Msg("schedule new job")

		err := fs.fk.Push(job)
		if err != nil {
			return fmt.Errorf("error pushing Job %s for feed %d: %w", jobType, feed.ID, err)
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
