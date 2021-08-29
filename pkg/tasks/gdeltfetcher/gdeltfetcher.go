// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gdeltfetcher

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/gdelt"
	"github.com/SpecializedGeneralist/gdelt/events"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/sets"
	faktory "github.com/contribsys/faktory/client"
	"github.com/jackc/pgtype"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"time"
)

// GDELTFetcher implements the mechanism for periodically fetching GDELT
// Events, collecting each event's first news report URL,
// enriched with essential Event metadata.
type GDELTFetcher struct {
	conf config.GDELTFetcher
	db   *gorm.DB
	fk   *faktory.Client
	log  zerolog.Logger
}

// New creates a new GDELTFetcher.
func New(conf config.GDELTFetcher, db *gorm.DB, fk *faktory.Client) *GDELTFetcher {
	return &GDELTFetcher{
		conf: conf,
		db:   db,
		fk:   fk,
		log:  log.Logger.Level(zerolog.Level(conf.LogLevel)),
	}
}

// Run starts the GDELT fetching process.
//
// This function should ideally run forever, unless an error is encountered
// or the context is done.
func (gf *GDELTFetcher) Run(ctx context.Context) (err error) {
	gf.log.Info().Msg("GDELT fetching starts")

Loop:
	for {
		gf.log.Info().Msg("fetching and processing events")
		err = gf.fetchAndProcessEvents(ctx)
		if err != nil {
			break
		}

		gf.log.Info().Msgf("waiting %s", gf.conf.TimeInterval)
		select {
		case <-time.After(gf.conf.TimeInterval):
		case <-ctx.Done():
			gf.log.Warn().Msg("context done")
			break Loop
		}
	}

	if err != nil {
		gf.log.Err(err).Msg("GDELT fetching ends with error")
		return err
	}

	gf.log.Info().Msg("GDELT fetching ends")
	return nil
}

func (gf *GDELTFetcher) fetchAndProcessEvents(ctx context.Context) error {
	evs, err := gdelt.GetLatestEvents()
	if err != nil {
		return fmt.Errorf("error fetching latest GDELT events: %w", err)
	}

	js := jobscheduler.New()

	err = gf.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		visitedURLs := sets.NewStringSetWithSize(len(evs))

		for _, ev := range evs {
			if visitedURLs.Has(ev.SourceURL) {
				continue
			}
			visitedURLs.Add(ev.SourceURL)

			err = gf.processEvent(tx, ev, js)
			if err != nil {
				return err
			}
		}
		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsWithClientAndDeletePendingJobs(gf.fk, gf.db)
}

func (gf *GDELTFetcher) processEvent(tx *gorm.DB, ev *events.Event, js *jobscheduler.JobScheduler) error {
	logger := gf.log.With().Uint64("GlobalEventID", ev.GlobalEventID).Logger()

	if len(ev.SourceURL) == 0 {
		logger.Debug().Msg("no source URL: skipping event")
		return nil
	}

	if !gf.eventRootCodeIsAllowed(ev.EventRootCode) {
		logger.Debug().Msgf("event root code %#v is not allowed: skipping event", ev.EventRootCode)
		return nil
	}

	webResource, err := findWebResource(tx, ev.SourceURL)
	if err != nil {
		return err
	}

	gdeltEvent, err := newGDELTEvent(ev)
	if err != nil {
		return err
	}

	if webResource != nil {
		logger = logger.With().Uint("WebResource", webResource.ID).Logger()

		if webResource.GDELTEvent != nil {
			logger.Debug().Uint("GDELTEvent", webResource.GDELTEvent.ID).Msg("a GDELT event already exists")
			return nil
		}

		gdeltEvent.WebResourceID = webResource.ID
		return createGDELTEvent(tx, logger, gdeltEvent)
	}

	webResource = &models.WebResource{
		URL:        ev.SourceURL,
		GDELTEvent: gdeltEvent,
	}

	res := tx.Create(webResource)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("WebResource and GDELTEvent creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating WebResource: %w", res.Error)
	}
	return js.AddJobs(gf.conf.NewWebResourceJobs, webResource.ID)
}

func createGDELTEvent(tx *gorm.DB, logger zerolog.Logger, ge *models.GDELTEvent) error {
	res := tx.Create(ge)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("GDELTEvent creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating GDELTEvent: %w", res.Error)
	}
	return nil
}

func findWebResource(tx *gorm.DB, url string) (*models.WebResource, error) {
	var webResource *models.WebResource
	result := tx.Joins("GDELTEvent").Limit(1).Find(&webResource, "url = ?", url)
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching WebResource by URL %#v: %w", url, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return webResource, nil
}

func (gf *GDELTFetcher) eventRootCodeIsAllowed(eventRootCode string) bool {
	if len(gf.conf.EventRootCodeWhitelist) == 0 {
		return true
	}
	for _, code := range gf.conf.EventRootCodeWhitelist {
		if code == eventRootCode {
			return true
		}
	}
	return false
}

func newGDELTEvent(ev *events.Event) (*models.GDELTEvent, error) {
	dateAddedTime, err := ev.DateAddedTime()
	if err != nil {
		return nil, fmt.Errorf("error getting DateAdded: %w", err)
	}

	countryCode, err := ev.ActionGeo.CountryCodeISO31661()
	if err != nil {
		return nil, fmt.Errorf("error getting country code: %w", err)
	}

	gdeltEvent := &models.GDELTEvent{
		GlobalEventID: uint(ev.GlobalEventID),
		DateAdded:     dateAddedTime,
		LocationType:  makeNullString(ev.ActionGeo.Type.String()),
		LocationName:  makeNullString(ev.ActionGeo.Fullname),
		CountryCode:   makeNullString(countryCode),
		Coordinates:   makePointCoordinates(ev.ActionGeo),
	}
	err = gdeltEvent.EventCategories.Set(ev.AllCameoEventCodes())
	if err != nil {
		return nil, fmt.Errorf("error setting EventCategories: %w", err)
	}
	return gdeltEvent, nil
}

func makePointCoordinates(g events.GeoData) pgtype.Point {
	if !g.Lat.Valid || !g.Long.Valid {
		return pgtype.Point{
			P:      pgtype.Vec2{X: 0, Y: 0},
			Status: pgtype.Null,
		}
	}
	return pgtype.Point{
		P:      pgtype.Vec2{X: g.Long.Float64, Y: g.Lat.Float64},
		Status: pgtype.Present,
	}
}

func makeNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{Valid: false, String: ""}
	}
	return sql.NullString{Valid: true, String: s}
}
