// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zeroshotclassification

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/workerpool"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// Classify starts the zero-shot classification workers.
func Classify(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start zero-shot classification")

	workers := make([]*Worker, config.ZeroShotClassification.NumWorkers)
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
			_ = workers[workerID].spagoGateway.Close()
		}
	}()

	wp := workerpool.New(config.ZeroShotClassification.NumWorkers)

	deliveries, consumerTag, err := rmq.Consume(
		config.ZeroShotClassification.SubQueueName,
		config.ZeroShotClassification.SubRoutingKey,
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

	logger.Info().Msg("stop zero-shot classification")
	return nil
}
