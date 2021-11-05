// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package textclassifier

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/textclassification"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"strings"
)

// TextClassifier implements a Faktory worker for classifying existing
// WebArticles with a generic text classifier.
type TextClassifier struct {
	basemodelworker.Worker
	// A custom function can be assigned for deciding whether to schedule or
	// not the configured next jobs upon completion
	//
	// The default value is DefaultShouldScheduleNextJobs.
	ShouldScheduleNextJobs ShouldScheduleNextJobsFn
	conf                   config.TextClassifier
}

// ShouldScheduleNextJobsFn is a function which returns a boolean flag
// that indicates whether a TextClassifier job should proceed scheduling
// the next configured jobs upon successful completion.
//
// Arguments:
// 	tx: the Gorm transaction instance created for the current job.
// 	    It can be used for getting data from the DB if needed; otherwise
//	    it can be ignored.
// 	wa: the WebArticle that has been classified. wa.TextClasses contains the
//      classification's results (depending on the implementation, it might
//      be empty). This value MUST NOT be modified.
//
// Returned values:
// 	- If true is returned with no error, the configured
//	  config.TextClassifier.ProcessedWebArticleJobs will be scheduled.
// 	- If false is returned with no error, no new jobs will be scheduled.
// 	- If the returned error is not nil, the boolean value is ignored and
// 	  the whole job will be aborted.
type ShouldScheduleNextJobsFn func(tx *gorm.DB, wa *models.WebArticle) (bool, error)

// New creates a new TextClassifier.
func New(
	conf config.TextClassifier,
	db *gorm.DB,
	fk *faktory_worker.Manager,
) *TextClassifier {
	tc := &TextClassifier{
		conf:                   conf,
		ShouldScheduleNextJobs: DefaultShouldScheduleNextJobs,
	}

	tc.Worker = basemodelworker.Worker{
		Name:        "TextClassifier",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Queues:      conf.Queues,
		Perform:     tc.perform,
	}
	return tc
}

func (tc *TextClassifier) perform(ctx context.Context, webArticleID uint) error {
	tx := tc.DB.WithContext(ctx)
	wa, err := getWebArticle(tx, webArticleID)
	if err != nil {
		return err
	}

	classes, err := tc.processWebArticle(ctx, wa)
	if err != nil {
		return err
	}

	js := jobscheduler.New()
	err = tx.Transaction(func(tx *gorm.DB) error {
		if len(classes) > 0 {
			res := tx.Create(&classes)
			if res.Error != nil {
				return fmt.Errorf("error saving new TextClasses: %w", res.Error)
			}
		}

		wa.TextClasses = classes
		shouldSchedule, err := tc.ShouldScheduleNextJobs(tx, wa)
		if err != nil {
			return err
		}
		if !shouldSchedule {
			return nil
		}

		err = js.AddJobs(tc.conf.ProcessedWebArticleJobs, wa.ID)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, tc.DB)
}

func getWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Preload("TextClasses").First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (tc *TextClassifier) processWebArticle(
	ctx context.Context,
	wa *models.WebArticle,
) ([]models.TextClass, error) {
	logger := tc.Log.With().Uint("WebArticle", wa.ID).Logger()

	if len(wa.TextClasses) > 0 {
		logger.Warn().Msg("this WebArticle already has TextClasses")
		return nil, nil
	}

	title := strings.TrimSpace(wa.Title)
	if wa.TranslatedTitle.Valid {
		title = strings.TrimSpace(wa.TranslatedTitle.String)
	}

	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return nil, nil
	}

	classifierConn, err := grpcconn.Dial(ctx, tc.conf.ClassifierServer)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := classifierConn.Close(); err != nil {
			tc.Log.Err(err).Msg("error closing classifier connection")
		}
	}()
	classifierClient := textclassification.NewClassifierClient(classifierConn)

	req := &textclassification.ClassifyTextRequest{Text: title}
	reply, err := classifierClient.ClassifyText(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ClassifyText request error: %w", err)
	}

	classes := make([]models.TextClass, len(reply.Classes))
	for i, repClass := range reply.Classes {
		classes[i] = models.TextClass{
			WebArticleID: wa.ID,
			Type:         repClass.Type,
			Label:        repClass.Label,
			Confidence:   repClass.Confidence,
		}
	}

	return classes, nil
}

// DefaultShouldScheduleNextJobs is the default implementation of
// TextClassifier.ShouldScheduleNextJobs.
//
// It always returns true and no error.
func DefaultShouldScheduleNextJobs(*gorm.DB, *models.WebArticle) (bool, error) {
	return true, nil
}
