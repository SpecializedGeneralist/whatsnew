// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package translator

import (
	"context"
	"database/sql"
	"fmt"
	translatorapi "github.com/SpecializedGeneralist/translator/pkg/api"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/sets"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

// Translator implements a Faktory worker for classifying existing
// WebArticles with spaGO BART zero-shot classification service.
type Translator struct {
	basemodelworker.Worker
	conf              config.Translator
	languageWhitelist sets.StringSet
	translatorClient  translatorapi.ApiClient
}

// New creates a new Translator.
func New(conf config.Translator, db *gorm.DB, translatorConn *grpc.ClientConn, fk *faktory_worker.Manager) *Translator {
	t := &Translator{
		conf:              conf,
		languageWhitelist: sets.NewStringSetWithElements(conf.LanguageWhitelist...),
		translatorClient:  translatorapi.NewApiClient(translatorConn),
	}

	t.Worker = basemodelworker.Worker{
		Name:        "Translator",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Queues:      conf.Queues,
		Perform:     t.perform,
	}
	return t
}

func (t *Translator) perform(ctx context.Context, webArticleID uint) error {
	js := jobscheduler.New()

	err := t.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wa, err := getLockedWebArticle(tx, webArticleID)
		if err != nil {
			return err
		}

		err = t.processWebArticle(ctx, tx, wa, js)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, t.DB)
}

func getLockedWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (t *Translator) processWebArticle(ctx context.Context, tx *gorm.DB, wa *models.WebArticle, js *jobscheduler.JobScheduler) error {
	logger := t.Log.With().Uint("WebArticle", wa.ID).Logger()

	if wa.TranslatedTitle.Valid && wa.TranslationLanguage.Valid {
		logger.Warn().Msg("this WebArticle already has a translated title")
		return nil
	}

	title := strings.TrimSpace(wa.Title)
	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return nil
	}

	if t.languageWhitelist.Has(wa.Language) {
		err := t.translateWebArticleTitle(ctx, tx, wa, title)
		if err != nil {
			return err
		}
	}

	return js.AddJobs(t.conf.ProcessedWebArticleJobs, wa.ID)
}

func (t *Translator) translateWebArticleTitle(ctx context.Context, tx *gorm.DB, wa *models.WebArticle, title string) error {
	resp, err := t.translatorClient.TranslateText(ctx, &translatorapi.TranslateTextRequest{
		TranslateTextInput: &translatorapi.TranslateTextInput{
			SourceLanguage: wa.Language,
			TargetLanguage: t.conf.TargetLanguage,
			Text:           title,
		},
	})
	if err != nil {
		return fmt.Errorf("TranslateText error: %w", err)
	}
	if resp.Errors != nil && len(resp.Errors.Value) > 0 {
		return fmt.Errorf("TranslateText responded with errors; first message: %s", resp.Errors.Value[0].Message)
	}

	translatedTitle := strings.TrimSpace(resp.Data.TranslatedText)
	if len(translatedTitle) == 0 {
		return fmt.Errorf("the title translation is empty")
	}

	wa.TranslatedTitle = sql.NullString{String: translatedTitle, Valid: true}
	wa.TranslationLanguage = sql.NullString{String: t.conf.TargetLanguage, Valid: true}

	res := tx.Save(wa)
	if res.Error != nil {
		return fmt.Errorf("error saving WebArticle with translated title: %w", res.Error)
	}
	return nil
}
