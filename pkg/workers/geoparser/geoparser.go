// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geoparser

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cliff"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"strings"
)

// GeoParser implements a Faktory worker for extracting geo-political
// entities from WebArticles.
type GeoParser struct {
	basemodelworker.Worker
	conf  config.GeoParser
	cliff *cliff.Client
}

// New creates a new GeoParser.
func New(
	conf config.GeoParser,
	db *gorm.DB,
	fk *faktory_worker.Manager,
) *GeoParser {
	gp := &GeoParser{
		conf:  conf,
		cliff: cliff.NewClient(conf.CliffURI),
	}
	gp.Worker = basemodelworker.Worker{
		Name:        "GeoParser",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Queues:      conf.Queues,
		Perform:     gp.perform,
	}
	return gp
}

func (gp *GeoParser) perform(ctx context.Context, webArticleID uint) error {
	tx := gp.DB.WithContext(ctx)

	wa, err := getWebArticle(tx, webArticleID)
	if err != nil {
		return err
	}

	countryOk, err := gp.processWebArticle(ctx, wa)
	if err != nil {
		return err
	}

	js := jobscheduler.New()
	err = tx.Transaction(func(tx *gorm.DB) error {
		if countryOk {
			err := models.OptimisticSave(tx, wa)
			if err != nil {
				return fmt.Errorf("error saving WebArticle: %w", err)
			}
		}

		err := js.AddJobs(gp.conf.ProcessedWebArticleJobs, wa.ID)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, gp.DB)
}

func getWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (gp *GeoParser) processWebArticle(
	ctx context.Context,
	wa *models.WebArticle,
) (bool, error) {
	logger := gp.Log.With().Uint("WebArticle", wa.ID).Logger()

	if wa.CountryCode.Valid {
		logger.Warn().Msg("this WebArticle already has a country code assigned")
		return false, nil
	}

	return gp.extractAndStoreCountry(ctx, wa)
}

func (gp *GeoParser) extractAndStoreCountry(
	ctx context.Context,
	wa *models.WebArticle,
) (bool, error) {
	logger := gp.Log.With().Uint("WebArticle", wa.ID).Logger()

	textOK, text, lang := chooseText(wa)
	if !textOK {
		logger.Debug().Msg("no text to parse")
		return false, nil
	}

	countryOK, code, err := gp.extractCountry(ctx, text, lang)
	if err != nil {
		return false, err
	}
	if !countryOK {
		logger.Debug().Msg("no country found")
		return false, nil
	}

	wa.CountryCode = sql.NullString{String: code, Valid: true}
	return true, nil
}

var languages = map[string]cliff.Language{
	"de": cliff.German,
	"es": cliff.Spanish,
	"en": cliff.English,
}

func chooseText(wa *models.WebArticle) (bool, string, cliff.Language) {
	lang, langOK := languages[wa.Language]
	text := strings.TrimSpace(wa.Title)
	if langOK && len(text) > 0 {
		return true, text, lang
	}

	if wa.TranslationLanguage.Valid && wa.TranslatedTitle.Valid {
		lang, langOK = languages[wa.TranslationLanguage.String]
		text = strings.TrimSpace(wa.TranslatedTitle.String)
		if langOK && len(text) > 0 {
			return true, text, lang
		}
	}

	return false, "", cliff.English
}

func (gp *GeoParser) extractCountry(
	ctx context.Context,
	text string,
	lang cliff.Language,
) (bool, string, error) {
	// Try without demonyms.
	pt, err := gp.cliff.ParseText(ctx, text, false, lang)
	if err != nil {
		return false, "", err
	}
	locs := pt.Results.Places.Focus.AllLocations()
	if len(locs) > 0 {
		return true, bestLocationCountryCode(locs), nil
	}

	// Otherwise, re-try with demonyms.
	pt, err = gp.cliff.ParseText(ctx, text, true, lang)
	if err != nil {
		return false, "", err
	}
	locs = pt.Results.Places.Focus.AllLocations()
	if len(locs) > 0 {
		return true, bestLocationCountryCode(locs), nil
	}
	return false, "", nil
}

func bestLocationCountryCode(locs []cliff.Location) string {
	best := locs[0]
	for _, loc := range locs[1:] {
		if loc.Score > best.Score {
			best = loc
		}
	}
	return best.CountryCode
}
