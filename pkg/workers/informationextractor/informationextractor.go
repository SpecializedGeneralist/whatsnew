// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package informationextractor

import (
	"context"
	"errors"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	bertgrpcapi "github.com/nlpodyssey/spago/pkg/nlp/transformers/bert/grpcapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"strings"
)

// InformationExtractor implements a Faktory worker for extracting information
// from WebArticles using spaGO BERT Question Answering service.
type InformationExtractor struct {
	basemodelworker.Worker
	conf config.InformationExtractor
}

// New creates a new InformationExtractor.
func New(
	conf config.InformationExtractor,
	db *gorm.DB,
	fk *faktory_worker.Manager,
) *InformationExtractor {
	ie := &InformationExtractor{
		conf: conf,
	}

	ie.Worker = basemodelworker.Worker{
		Name:        "InformationExtractor",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Queues:      conf.Queues,
		Perform:     ie.perform,
	}
	return ie
}

var errSkip = errors.New("skip")

func (ie *InformationExtractor) perform(ctx context.Context, webArticleID uint) error {
	tx := ie.DB.WithContext(ctx)

	wa, err := getWebArticle(tx, webArticleID)
	if err != nil {
		return err
	}

	infos, err := ie.processWebArticle(ctx, tx, wa)
	if errors.Is(err, errSkip) {
		return nil
	}
	if err != nil {
		return err
	}

	js := jobscheduler.New()
	err = tx.Transaction(func(tx *gorm.DB) error {
		if len(infos) > 0 {
			res := tx.Create(&infos)
			if res.Error != nil {
				return fmt.Errorf("error saving new ExtractedInfo models: %w", res.Error)
			}
		}

		return js.AddJobsAndCreatePendingJobs(tx, ie.conf.ProcessedWebArticleJobs, wa.ID)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, ie.DB)
}

func getWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Preload("ExtractedInfos").First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (ie *InformationExtractor) processWebArticle(
	ctx context.Context,
	tx *gorm.DB,
	wa *models.WebArticle,
) ([]*models.ExtractedInfo, error) {
	logger := ie.Log.With().Uint("WebArticle", wa.ID).Logger()

	if len(wa.ExtractedInfos) > 0 {
		logger.Warn().Msg("this WebArticle already has extracted info")
		return nil, errSkip
	}

	title := strings.TrimSpace(wa.Title)
	if wa.TranslatedTitle.Valid {
		title = strings.TrimSpace(wa.TranslatedTitle.String)
	}

	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return nil, errSkip
	}

	return ie.extractAndSaveInfo(ctx, tx, wa, title)
}

func (ie *InformationExtractor) extractAndSaveInfo(
	ctx context.Context,
	tx *gorm.DB,
	wa *models.WebArticle,
	title string,
) ([]*models.ExtractedInfo, error) {
	rules, err := ie.getRules(tx)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}

	infos := make([]*models.ExtractedInfo, 0, len(rules))

	for _, rule := range rules {
		ans, err := ie.getBestAnswer(ctx, title, rule.Question)
		if err != nil {
			return nil, err
		}
		if ans == nil {
			continue
		}
		confidence := float32(ans.Confidence)
		logger := ie.Log.With().Uint("ruleID", rule.ID).Str("answer", ans.Text).
			Float32("confidence", confidence).Logger()

		if confidence < rule.Threshold {
			logger.Debug().Msg("answer confidence below threshold")
			continue
		}

		if !rule.AnswerRegexp.MatchString(ans.Text) {
			logger.Debug().Msg("no regexp match")
			continue
		}

		infos = append(infos, &models.ExtractedInfo{
			WebArticleID:         wa.ID,
			InfoExtractionRuleID: rule.ID,
			Text:                 ans.Text,
			Confidence:           confidence,
		})
	}

	return infos, nil
}

func (ie *InformationExtractor) getBestAnswer(
	ctx context.Context,
	passage,
	question string,
) (*bertgrpcapi.Answer, error) {
	bertConn, err := grpcconn.Dial(ctx, ie.conf.SpagoBERTServer)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := bertConn.Close(); err != nil {
			ie.Log.Err(err).Msg("error closing BERT connection")
		}
	}()
	bertClient := bertgrpcapi.NewBERTClient(bertConn)

	reply, err := bertClient.Answer(ctx, &bertgrpcapi.AnswerRequest{
		Passage:  strings.ToLower(passage),
		Question: strings.ToLower(question),
	})
	if err != nil {
		return nil, fmt.Errorf("BERT Q/A Answer error: %w", err)
	}
	if len(reply.Answers) == 0 {
		return nil, nil
	}
	return reply.Answers[0], nil
}

func (ie *InformationExtractor) getRules(tx *gorm.DB) ([]models.InfoExtractionRule, error) {
	var rules []models.InfoExtractionRule
	res := tx.Find(&rules, "enabled")
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching InfoExtractionRules: %w", res.Error)
	}
	return rules, nil
}
