// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feedsfetching

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/workerpool"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"time"
)

func FetchFeeds(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start fetching feeds")

	workers := make([]*Worker, config.FeedsFetching.NumWorkers)
	for workerID := range workers {
		workers[workerID] = NewWorker(
			config, db, rmq,
			logger.With().Int("workerID", workerID).Logger(),
		)
	}

	wp := workerpool.New(config.FeedsFetching.NumWorkers)

	go runProducer(config, db, wp, logger)

	wp.Run(func(workerID int, jobData interface{}) {
		workers[workerID].Do(jobData.(uint))
	})

	logger.Info().Msg("stop fetching feeds")
	return nil
}

func runProducer(
	config configuration.Configuration,
	db *gorm.DB,
	wp *workerpool.WorkerPool,
	logger zerolog.Logger,
) {
	for {
		logger.Info().Msg("producer: start finding feeds in batch")

		var results []*struct{ ID uint }
		result := db.Model(&models.Feed{}).Order("last_retrieved_at, id").FindInBatches(
			&results, 100,
			func(_ *gorm.DB, batch int) error {
				for _, result := range results {
					wp.PublishJobData(result.ID)
				}
				return nil
			},
		)

		sleepingDuration := config.FeedsFetching.SleepingTime
		if result.Error != nil {
			logger.Err(result.Error).Msg("producer: FindInBatches error")
			sleepingDuration = 10 * time.Second
		}

		logger.Info().Msgf("producer: sleeping for %s...", sleepingDuration)
		time.Sleep(sleepingDuration)
	}
}
