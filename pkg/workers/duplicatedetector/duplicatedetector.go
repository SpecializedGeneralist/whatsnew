// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package duplicatedetector

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/hnswclient"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// DuplicateDetector implements a Faktory worker for performing near-duplicate
// detection over existing WebArticles.
type DuplicateDetector struct {
	basemodelworker.Worker
	conf       config.DuplicateDetector
	hnswClient *hnswclient.Client
}

const day = 24 * time.Hour

// New creates a new WebScraper.
func New(conf config.DuplicateDetector, db *gorm.DB, hnswClient *hnswclient.Client, fk *faktory_worker.Manager) *DuplicateDetector {
	v := &DuplicateDetector{
		conf:       conf,
		hnswClient: hnswClient,
	}
	v.Worker = basemodelworker.Worker{
		Name:        "DuplicateDetector",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: 1,
		Perform:     v.perform,
	}
	return v
}

func (dd *DuplicateDetector) perform(ctx context.Context, webArticleID uint) error {
	js := jobscheduler.New()

	err := dd.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wa, err := getLockedWebArticle(tx, webArticleID)
		if err != nil {
			return err
		}

		err = dd.processWebArticle(ctx, tx, wa, js)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, dd.DB)
}

func getLockedWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Vector").
		Preload("SimilarityInfo").
		First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (dd *DuplicateDetector) processWebArticle(ctx context.Context, tx *gorm.DB, wa *models.WebArticle, js *jobscheduler.JobScheduler) error {
	logger := dd.Log.With().Uint("WebArticle", wa.ID).Logger()

	if wa.SimilarityInfo != nil {
		logger.Warn().Msg("SimilarityInfo is already present on this WebArticle")
		return nil
	}
	if wa.Vector == nil {
		logger.Warn().Msg("this WebArticle does not have a vector")
		return nil
	}

	hit, err := dd.findSimilarHit(ctx, wa)
	if err != nil {
		return err
	}

	si := newSimilarityInfo(wa, hit)
	res := tx.Create(si)
	if res.Error != nil {
		return fmt.Errorf("error creating SimilarityInfo: %w", res.Error)
	}

	if hit != nil {
		return js.AddJobs(dd.conf.NonDuplicateWebArticleJobs, wa.ID)
	}
	return js.AddJobs(dd.conf.DuplicateWebArticleJobs, wa.ID)
}

func newSimilarityInfo(wa *models.WebArticle, hit *hnswclient.Hit) *models.SimilarityInfo {
	si := &models.SimilarityInfo{WebArticleID: wa.ID}
	if hit != nil {
		si.ParentID = &hit.ID
		si.Distance = &hit.Distance
	}
	return si
}

func (dd *DuplicateDetector) findSimilarHit(ctx context.Context, wa *models.WebArticle) (*hnswclient.Hit, error) {
	vector, err := wa.Vector.DataAsFloat32Slice()
	if err != nil {
		return nil, err
	}

	hits, err := dd.hnswClient.SearchKNN(ctx, hnswclient.SearchParams{
		From:              wa.PublishDate.Add(-time.Duration(dd.conf.TimeframeDays) * day),
		To:                wa.PublishDate,
		Vector:            vector,
		DistanceThreshold: dd.conf.DistanceThreshold,
	})
	if err != nil {
		return nil, err
	}

	for _, hit := range hits {
		// Skip the ID of the WebArticle itself.
		// To prevent mutual similarity, always avoid marking a WebArticle
		// with smaller ID as similar to a WebArticle with bigger ID.
		if hit.ID < wa.ID {
			return &hit, nil
		}
	}
	return nil, nil
}
