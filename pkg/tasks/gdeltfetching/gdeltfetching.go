// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gdeltfetching

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// FetchGDELT periodically fetches GDELT's Events, collecting each event's
// first news report URL, enriched with essential Event metadata.
func FetchGDELT(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) error {
	logger.Info().Msg("start fetching GDELT events")

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	w := NewWorker(config, db, rmq, logger)

	for done := false; !done; {
		logger.Info().Msg("fetching GDELT data")
		err := w.Do()

		var sleepingTime time.Duration
		if err == nil {
			sleepingTime = config.GDELTFetching.SleepingTimeSeconds
			logger.Info().Msgf("sleeping for %d seconds...", sleepingTime)
		} else {
			sleepingTime = 30
			logger.Err(err).Msgf("an error occurred - retrying in %d seconds...", sleepingTime)
		}

		select {
		case <-time.After(sleepingTime * time.Second):
		case s := <-termChan:
			logger.Info().Msgf("signal received: %v", s)
			done = true
		}
	}
	logger.Info().Msg("stop fetching GDELT events")

	signal.Stop(termChan)
	close(termChan)
	return nil
}
