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
		Queues:      conf.Queues,
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

	templates, err := zsc.getHypotheses(tx)
	if err != nil {
		return err
	}

	var classes []*models.ZeroShotClass
	for _, template := range templates {
		newClasses, err := zsc.classify(ctx, wa.ID, title, template)
		if err != nil {
			return err
		}
		classes = append(classes, newClasses...)
	}

	if len(classes) > 0 {
		res := tx.Create(classes)
		if res.Error != nil {
			return fmt.Errorf("error saving new ZeroShotClasses: %w", res.Error)
		}
	}

	return js.AddJobs(zsc.conf.ProcessedWebArticleJobs, wa.ID)
}

func (zsc *ZeroShotClassifier) getHypotheses(tx *gorm.DB) ([]models.ZeroShotHypothesisTemplate, error) {
	var templates []models.ZeroShotHypothesisTemplate
	res := tx.Preload("Labels", "enabled").Find(&templates, "enabled")
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching ZeroShotHypothesisTemplates: %w", res.Error)
	}
	return templates, nil
}

func (zsc *ZeroShotClassifier) classify(ctx context.Context, webArticleID uint, text string, template models.ZeroShotHypothesisTemplate) ([]*models.ZeroShotClass, error) {
	if len(template.Labels) == 0 {
		// This happens if a template has no enabled labels
		return nil, nil
	}

	possibleLabels := make([]string, len(template.Labels))
	labelToID := make(map[string]uint, len(template.Labels))
	for i, l := range template.Labels {
		possibleLabels[i] = l.Text
		if _, exists := labelToID[l.Text]; exists {
			return nil, fmt.Errorf("duplicate label %#v for hypothesis template %d", l.Text, template.ID)
		}
		labelToID[l.Text] = l.ID
	}

	reply, err := zsc.bartClient.ClassifyNLI(ctx, &grpcapi.ClassifyNLIRequest{
		Text:               text,
		HypothesisTemplate: template.Text,
		PossibleLabels:     possibleLabels,
		MultiClass:         template.MultiClass,
	})
	if err != nil {
		return nil, fmt.Errorf("BART ClassifyNLI error: %w", err)
	}
	distribution := reply.GetDistribution()
	if len(distribution) == 0 {
		return nil, fmt.Errorf("BART ClassifyNLI returned an empty distribution")
	}

	classes := make([]*models.ZeroShotClass, len(distribution))
	for i, pair := range distribution {
		labelID, ok := labelToID[pair.Class]
		if !ok {
			return nil, fmt.Errorf("ClassifyNLI returned an unknown class: %#v", pair.Class)
		}
		classes[i] = &models.ZeroShotClass{
			WebArticleID:                 webArticleID,
			ZeroShotHypothesisLabelID:    labelID,
			ZeroShotHypothesisTemplateID: template.ID,
			Best:                         i == 0,
			Confidence:                   float32(pair.Confidence),
		}
	}
	return classes, nil
}
