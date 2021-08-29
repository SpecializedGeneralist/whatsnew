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
	// A custom function can be assigned for selecting the topmost similar
	// hit among all HNSW KNN search results.
	//
	// The default value is DefaultSelectTopHit.
	SelectTopHit SelectTopHitFn
	basemodelworker.Worker
	conf       config.DuplicateDetector
	hnswClient *hnswclient.Client
}

// SelectTopHitFn is a function type for selecting the top similar
// entry among all KNN search hits.
//
// Arguments:
// 	tx: the Gorm transaction instance created for the current job.
// 	    It can be used for getting data from the DB in order to implement
//	    specific filtering criteria; otherwise it can be ignored.
// 	wa: the WebArticle whose Vector (already preloaded) was used
// 	    for HNSW KNN Search, obtaining the "hits".
//	    This value MUST NOT be modified.
// 	hits: the value returned from hnswclient.Client.SearchKNN().
//	      Please note that, according to the default implementation of other
//	      workers, it might always include the ID of the WebArticle "wa" itself.
//	      This value MUST NOT be modified.
//
// Returned values:
// 	- If a non-nil *Hit is returned with no error, "wa" will be considered a
// 	  duplicate of the "parent" WebArticle identified by Hit.ID. The Hit's ID
// 	  and Distance will be stored in the new SimilarityInfo model associated
// 	  to "wa", as SimilarityInfo.ParentID and SimilarityInfo.Distance
//	  respectively.
// 	- If a nil *Hit is returned with no error, "wa" is not considered a
//	  duplicate of another WebArticle (has no parent). The new SimilarityInfo
//	  model associated to "wa" will have the neither ParentID nor Distance.
// 	- If the returned error is not nil, the *Hit value will be ignored and
// 	  the whole job will be aborted.
type SelectTopHitFn func(tx *gorm.DB, wa *models.WebArticle, hits hnswclient.Hits) (*hnswclient.Hit, error)

const day = 24 * time.Hour

// New creates a new WebScraper.
func New(conf config.DuplicateDetector, db *gorm.DB, hnswClient *hnswclient.Client, fk *faktory_worker.Manager) *DuplicateDetector {
	v := &DuplicateDetector{
		SelectTopHit: DefaultSelectTopHit,
		conf:         conf,
		hnswClient:   hnswClient,
	}
	v.Worker = basemodelworker.Worker{
		Name:        "DuplicateDetector",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: 1,
		Queues:      conf.Queues,
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

	hit, err := dd.findSimilarHit(ctx, tx, wa)
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

func (dd *DuplicateDetector) findSimilarHit(ctx context.Context, tx *gorm.DB, wa *models.WebArticle) (*hnswclient.Hit, error) {
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

	return dd.SelectTopHit(tx, wa, hits)
}

// DefaultSelectTopHit is the default implementation of
// DuplicateDetector.SelectTopHit.
//
// It simply returns the first element among "hits", if any, obtained by
// skipping the ID of the WebArticle itself and also ignoring hits whose ID is
// larger than the ID of "wa" (this is done to prevent mutual similarity
// between WebArticles).
func DefaultSelectTopHit(_ *gorm.DB, wa *models.WebArticle, hits hnswclient.Hits) (*hnswclient.Hit, error) {
	for _, hit := range hits {
		if hit.ID < wa.ID {
			return &hit, nil
		}
	}
	return nil, nil
}
