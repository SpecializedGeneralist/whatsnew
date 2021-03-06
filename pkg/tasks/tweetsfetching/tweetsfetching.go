// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tweetsfetching

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/workerpool"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"time"
)

func FetchTweets(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start fetching tweets")

	workers := make([]*Worker, config.TweetsFetching.NumWorkers)
	for workerID := range workers {
		workers[workerID] = NewWorker(
			config, db.Session(&gorm.Session{}), rmq,
			logger.With().Int("workerID", workerID).Logger(),
		)
	}

	wp := workerpool.New(config.TweetsFetching.NumWorkers)

	go runProducer(config, db, wp, logger)

	wp.Run(func(workerID int, jobData interface{}) {
		workers[workerID].Do(jobData.(uint))
	})

	logger.Info().Msg("stop fetching tweets")
	return nil
}

func runProducer(
	config configuration.Configuration,
	db *gorm.DB,
	wp *workerpool.WorkerPool,
	logger zerolog.Logger,
) {
	for {
		logger.Info().Msg("producer: start finding twitter-sources in batch")

		var results []*models.TwitterSource
		result := db.Order("last_retrieved_at, id").FindInBatches(
			&results, 100,
			func(_ *gorm.DB, batch int) error {
				for _, result := range results {
					wp.PublishJobData(result.ID)
				}
				return nil
			},
		)

		sleepingDuration := config.TweetsFetching.SleepingTime
		if result.Error != nil {
			logger.Err(result.Error).Msg("producer: FindInBatches error")
			sleepingDuration = 10 * time.Second
		}

		logger.Info().Msgf("producer: sleeping for %s...", sleepingDuration)
		time.Sleep(sleepingDuration)
	}
}
