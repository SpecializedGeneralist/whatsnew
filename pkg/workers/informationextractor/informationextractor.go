// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package informationextractor

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	bert_grpcapi "github.com/nlpodyssey/spago/pkg/nlp/transformers/bert/grpcapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"regexp"
	"strings"
)

// InformationExtractor implements a Faktory worker for extracting information
// fomr WebArticles using spaGO BERT Question Answering service.
type InformationExtractor struct {
	basemodelworker.Worker
	conf       config.InformationExtractor
	bertClient bert_grpcapi.BERTClient
}

// New creates a new InformationExtractor.
func New(conf config.InformationExtractor, db *gorm.DB, bertConn *grpc.ClientConn, fk *faktory_worker.Manager) *InformationExtractor {
	ie := &InformationExtractor{
		conf:       conf,
		bertClient: bert_grpcapi.NewBERTClient(bertConn),
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

func (ie *InformationExtractor) perform(ctx context.Context, webArticleID uint) error {
	js := jobscheduler.New()

	err := ie.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wa, err := getLockedWebArticle(tx, webArticleID)
		if err != nil {
			return err
		}

		err = ie.processWebArticle(ctx, tx, wa, js)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, ie.DB)
}

func getLockedWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("ExtractedInfos").First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (ie *InformationExtractor) processWebArticle(ctx context.Context, tx *gorm.DB, wa *models.WebArticle, js *jobscheduler.JobScheduler) error {
	logger := ie.Log.With().Uint("WebArticle", wa.ID).Logger()

	if len(wa.ExtractedInfos) > 0 {
		logger.Warn().Msg("this WebArticle already has extracted info")
		return nil
	}

	title := strings.TrimSpace(wa.Title)
	if wa.TranslatedTitle.Valid {
		title = strings.TrimSpace(wa.TranslatedTitle.String)
	}

	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return nil
	}

	err := ie.extractAndSaveInfo(ctx, tx, wa, title)
	if err != nil {
		return err
	}

	return js.AddJobs(ie.conf.ProcessedWebArticleJobs, wa.ID)
}

func (ie *InformationExtractor) extractAndSaveInfo(ctx context.Context, tx *gorm.DB, wa *models.WebArticle, title string) error {
	infos := make([]*models.ExtractedInfo, 0, len(ie.conf.Items))

	for _, item := range ie.conf.Items {
		ans, err := ie.getBestAnswer(ctx, title, item.Question)
		if err != nil {
			return err
		}
		if ans == nil {
			continue
		}
		confidence := float32(ans.Confidence)
		logger := ie.Log.With().Str("question", item.Question).Str("answer", ans.Text).Float32("confidence", confidence).Logger()

		if confidence < item.Threshold {
			logger.Trace().Float32("threshold", item.Threshold).Msg("answer confidence below threshold")
			continue
		}

		re := regexp.Regexp(item.AnswerRegexp)
		if !(&re).MatchString(ans.Text) {
			logger.Trace().Stringer("regexp", &re).Msg("no regexp match")
			continue
		}

		infos = append(infos, &models.ExtractedInfo{
			WebArticleID: wa.ID,
			Label:        item.Label,
			Text:         ans.Text,
			Confidence:   confidence,
		})
	}

	if len(infos) == 0 {
		return nil
	}

	res := tx.Save(infos)
	if res.Error != nil {
		return fmt.Errorf("error saving new ExtractedInfo models: %w", res.Error)
	}
	return nil
}

func (ie *InformationExtractor) getBestAnswer(ctx context.Context, passage, question string) (*bert_grpcapi.Answer, error) {
	reply, err := ie.bertClient.Answer(ctx, &bert_grpcapi.AnswerRequest{
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