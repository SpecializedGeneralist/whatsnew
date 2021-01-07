// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webscraping

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/workerpool"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

func ScrapeWeb(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start scraping Web Resources")

	workers := make([]*Worker, config.WebScraping.NumWorkers)
	for workerID := range workers {
		workers[workerID] = NewWorker(
			config, db, rmq,
			logger.With().Int("workerID", workerID).Logger(),
		)
	}

	wp := workerpool.New(config.WebScraping.NumWorkers)

	deliveries, consumerTag, err := rmq.Consume(
		config.WebScraping.SubQueueName,
		config.WebScraping.SubNewWebResourceRoutingKey,
	)
	if err != nil {
		return fmt.Errorf("starting consuming: %v", err)
	}

	go func() {
		for delivery := range deliveries {
			wp.PublishJobData(delivery)
		}
	}()

	wp.Run(func(workerID int, jobData interface{}) {
		workers[workerID].Do(jobData.(amqp.Delivery))
	})

	err = rmq.CancelConsumer(consumerTag)
	if err != nil {
		logger.Err(err).Msg("error canceling RabbitMQ channel")
	}

	logger.Info().Msg("stop web scraping")
	return nil
}
