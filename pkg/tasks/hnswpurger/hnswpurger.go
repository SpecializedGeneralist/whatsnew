// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hnswpurger

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/hnswclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

// HNSWPurger implements the mechanism for periodically checking for old
// PendingJobs and attempting their rescheduling.
type HNSWPurger struct {
	conf     config.HNSWPurger
	hnswConf config.HNSW
	log      zerolog.Logger
}

// New creates a new HNSWPurger.
func New(conf config.HNSWPurger, hnswConf config.HNSW) *HNSWPurger {
	return &HNSWPurger{
		conf:     conf,
		hnswConf: hnswConf,
		log:      log.Logger.Level(zerolog.Level(conf.LogLevel)),
	}
}

// Run starts the pending jobs' recovery process.
//
// This function should ideally run forever, unless an error is encountered
// or the context is done.
func (hp *HNSWPurger) Run(ctx context.Context) (err error) {
	hp.log.Info().Msg("HNSW purging task starts")

Loop:
	for {
		err = hp.purge(ctx)
		if err != nil {
			break
		}

		hp.log.Info().Msgf("waiting %s", hp.conf.TimeInterval)
		select {
		case <-time.After(hp.conf.TimeInterval):
		case <-ctx.Done():
			hp.log.Warn().Msg("context done")
			break Loop
		}
	}

	if err != nil {
		hp.log.Err(err).Msg("HNSW purging task ends with error")
		return err
	}

	hp.log.Info().Msg("HNSW purge ends")
	return nil
}

const day = 24 * time.Hour

func (hp *HNSWPurger) purge(ctx context.Context) error {
	hnswConn, err := grpcconn.Dial(ctx, hp.hnswConf.Server)
	if err != nil {
		return err
	}
	defer func() {
		if err := hnswConn.Close(); err != nil {
			hp.log.Err(err).Msg("error closing HNSW connection")
		}
	}()
	hnswClient := hnswclient.New(hnswConn, hp.hnswConf.Index)

	daysFromNow := time.Duration(hp.conf.DeleteIndicesOlderThanDays) * day
	upperTime := time.Now().UTC().Add(-daysFromNow)

	indices, err := hnswClient.IndicesOlderThan(ctx, upperTime)
	if err != nil {
		return err
	}

	hp.log.Info().Msgf("%d old indices to delete", len(indices))

	for _, index := range indices {
		hp.log.Info().Msgf("deleting index %#v", index)
		err := hnswClient.DeleteIndex(ctx, index)
		if err != nil {
			return err
		}
	}

	return hnswClient.FlushAllIndices(ctx)
}
