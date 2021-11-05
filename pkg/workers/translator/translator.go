// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package translator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	translatorapi "github.com/SpecializedGeneralist/translator/pkg/api"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/sets"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"strings"
)

// Translator implements a Faktory worker for classifying existing
// WebArticles with spaGO BART zero-shot classification service.
type Translator struct {
	basemodelworker.Worker
	conf              config.Translator
	languageWhitelist sets.StringSet
}

// New creates a new Translator.
func New(conf config.Translator, db *gorm.DB, fk *faktory_worker.Manager) *Translator {
	t := &Translator{
		conf:              conf,
		languageWhitelist: sets.NewStringSetWithElements(conf.LanguageWhitelist...),
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

var errSkip = errors.New("skip")

func (t *Translator) perform(ctx context.Context, webArticleID uint) error {
	tx := t.DB.WithContext(ctx)

	wa, err := getWebArticle(tx, webArticleID)
	if err != nil {
		return err
	}

	translationOk, err := t.processWebArticle(ctx, wa)
	if errors.Is(err, errSkip) {
		return nil
	}
	if err != nil {
		return err
	}

	js := jobscheduler.New()
	err = tx.Transaction(func(tx *gorm.DB) error {
		if translationOk {
			err := models.OptimisticSave(tx, wa)
			if err != nil {
				return fmt.Errorf("error saving WebArticle with translated title: %w", err)
			}
		}

		err := js.AddJobs(t.conf.ProcessedWebArticleJobs, wa.ID)
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

func getWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (t *Translator) processWebArticle(
	ctx context.Context,
	wa *models.WebArticle,
) (bool, error) {
	logger := t.Log.With().Uint("WebArticle", wa.ID).Logger()

	if wa.TranslatedTitle.Valid && wa.TranslationLanguage.Valid {
		logger.Warn().Msg("this WebArticle already has a translated title")
		return false, errSkip
	}

	title := strings.TrimSpace(wa.Title)
	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return false, errSkip
	}

	if !t.languageWhitelist.Has(wa.Language) {
		return false, nil
	}

	err := t.translateTitle(ctx, wa, title)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (t *Translator) translateTitle(ctx context.Context, wa *models.WebArticle, title string) error {
	translatorConn, err := grpcconn.Dial(ctx, t.conf.TranslatorServer)
	if err != nil {
		return err
	}
	defer func() {
		if err := translatorConn.Close(); err != nil {
			t.Log.Err(err).Msg("error closing translator connection")
		}
	}()
	translatorClient := translatorapi.NewApiClient(translatorConn)

	resp, err := translatorClient.TranslateText(ctx, &translatorapi.TranslateTextRequest{
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
	return nil
}
