// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package duplicatedetector

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
)

type CustomizeQueryFunc func(tx *gorm.DB, webArticle *models.WebArticle) (*gorm.DB, error)

func DefaultDetectDuplicates(
	config configuration.DuplicateDetectorConfiguration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	return DetectDuplicates(config, db, rmq, nil, logger)
}

func DetectDuplicates(
	config configuration.DuplicateDetectorConfiguration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	customizeQuery CustomizeQueryFunc,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start duplication detection")

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	deliveries, consumerTag, err := rmq.Consume(
		config.SubQueueName,
		config.SubRoutingKey,
	)
	if err != nil {
		return fmt.Errorf("starting consuming: %v", err)
	}

	worker := &Worker{
		config:         config,
		db:             db,
		rmq:            rmq,
		customizeQuery: customizeQuery,
		logger:         logger,
	}

	for done := false; !done; {
		select {
		case delivery := <-deliveries:
			worker.do(delivery)
		case s := <-termChan:
			logger.Info().Msgf("signal received: %v", s)
			done = true
		}
	}

	logger.Info().Msg("stop duplication detection")

	signal.Stop(termChan)
	close(termChan)

	err = rmq.CancelConsumer(consumerTag)
	if err != nil {
		logger.Err(err).Msg("error canceling RabbitMQ channel")
	}

	return nil
}
