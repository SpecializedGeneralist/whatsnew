// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vectorizer

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/workerpool"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

func Vectorize(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start articles vectorization")

	workers := make([]*Worker, config.Vectorizer.NumWorkers)
	for workerID := range workers {
		var err error
		workers[workerID], err = NewWorker(
			config, db, rmq,
			logger.With().Int("workerID", workerID).Logger(),
		)
		if err != nil {
			return err
		}
	}

	defer func() {
		for workerID := range workers {
			_ = workers[workerID].labseGateway.Close()
		}
	}()

	wp := workerpool.New(config.Vectorizer.NumWorkers)

	deliveries, consumerTag, err := rmq.Consume(
		config.Vectorizer.SubQueueName,
		config.Vectorizer.SubNewWebArticleRoutingKey,
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

	logger.Info().Msg("stop articles vectorization")
	return nil
}
