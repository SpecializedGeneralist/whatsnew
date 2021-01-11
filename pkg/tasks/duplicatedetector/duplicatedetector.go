// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package duplicatedetector

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
)

func DetectDuplicates(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start duplication detection")

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	deliveries, consumerTag, err := rmq.Consume(
		config.DuplicateDetector.SubQueueName,
		config.DuplicateDetector.SubRoutingKey,
	)
	if err != nil {
		return fmt.Errorf("starting consuming: %v", err)
	}

	worker := &Worker{
		config: config,
		db:     db,
		rmq:    rmq,
		logger: logger,
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
