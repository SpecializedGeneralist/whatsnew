// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zeroshotclassifier

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/nlpodyssey/spago/pkg/nlp/transformers/bart/server/grpcapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

// ZeroShotClassifier implements a Faktory worker for classifying existing
// WebArticles with spaGO BART zero-shot classification service.
type ZeroShotClassifier struct {
	basemodelworker.Worker
	conf       config.ZeroShotClassifier
	bartClient grpcapi.BARTClient
}

// New creates a new ZeroShotClassifier.
func New(conf config.ZeroShotClassifier, db *gorm.DB, bartConn *grpc.ClientConn, fk *faktory_worker.Manager) *ZeroShotClassifier {
	zsc := &ZeroShotClassifier{
		conf:       conf,
		bartClient: grpcapi.NewBARTClient(bartConn),
	}

	zsc.Worker = basemodelworker.Worker{
		Name:        "ZeroShotClassifier",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Perform:     zsc.perform,
	}
	return zsc
}

func (zsc *ZeroShotClassifier) perform(ctx context.Context, webArticleID uint) error {
	js := jobscheduler.New()

	err := zsc.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wa, err := getLockedWebArticle(tx, webArticleID)
		if err != nil {
			return err
		}

		err = zsc.processWebArticle(ctx, tx, wa, js)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, zsc.DB)
}

func getLockedWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("ZeroShotClasses").First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (zsc *ZeroShotClassifier) processWebArticle(ctx context.Context, tx *gorm.DB, wa *models.WebArticle, js *jobscheduler.JobScheduler) error {
	logger := zsc.Log.With().Uint("WebArticle", wa.ID).Logger()

	if len(wa.ZeroShotClasses) > 0 {
		logger.Warn().Msg("this WebArticle already has classes")
		return nil
	}

	title := strings.TrimSpace(wa.Title)
	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return nil
	}

	reply, err := zsc.bartClient.ClassifyNLI(ctx, &grpcapi.ClassifyNLIRequest{
		Text:               title,
		HypothesisTemplate: zsc.conf.HypothesisTemplate,
		PossibleLabels:     zsc.conf.PossibleLabels,
		MultiClass:         zsc.conf.MultiClass,
	})
	if err != nil {
		return fmt.Errorf("BART ClassifyNLI error: %w", err)
	}

	distribution := reply.GetDistribution()
	if len(distribution) == 0 {
		return fmt.Errorf("BART ClassifyNLI returned an empty distribution")
	}

	classes := make([]*models.ZeroShotClass, len(distribution))
	for i, pair := range distribution {
		classes[i] = &models.ZeroShotClass{
			WebArticleID: wa.ID,
			Class:        pair.Class,
			Confidence:   pair.Confidence,
			Best:         i == 0,
		}
	}

	res := tx.Create(classes)
	if res.Error != nil {
		return fmt.Errorf("error saving new ZeroShotClasses: %w", res.Error)
	}

	return js.AddJobs(zsc.conf.ClassifiedWebArticleJobs, wa.ID)
}
