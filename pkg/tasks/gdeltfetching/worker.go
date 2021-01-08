// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gdeltfetching

import (
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/jackc/pgtype"
	"github.com/rs/zerolog"
	"github.com/specializedgeneralist/gdelt"
	"github.com/specializedgeneralist/gdelt/events"
	"gorm.io/gorm"
)

type Worker struct {
	config configuration.Configuration
	db     *gorm.DB
	rmq    *rabbitmq.Client
	logger zerolog.Logger
}

func NewWorker(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) *Worker {
	return &Worker{
		config: config,
		db:     db,
		rmq:    rmq,
		logger: logger,
	}
}

func (g *Worker) Do() error {
	eventRecords, err := gdelt.GetLatestEvents()
	if err != nil {
		return fmt.Errorf("fetching events from CSV Zip reference: %v", err)
	}

	for _, record := range eventRecords {
		g.processEventRecord(record)
	}
	return nil
}

func (g *Worker) processEventRecord(event *events.Event) {
	logger := g.logger.With().Uint64("GlobalEventID", event.GlobalEventID).Logger()

	if len(event.SourceURL) == 0 {
		logger.Debug().Msg("skipping event without source URL")
		return
	}

	if !g.cameoTopLevelEventCodeIsWhitelisted(event) {
		logger.Debug().Str("EventRootCode", event.EventRootCode).
			Msg("CAMEO event root code is not whitelisted: skipping event")
		return
	}

	webResource, err := models.FindWebResourceByURL(g.db, event.SourceURL)
	if err != nil {
		logger.Err(err).Msg("error finding web resource by URL")
		return
	}

	if webResource != nil {
		gdeltEvent, err := g.createGDELTEventIfItDoesNotExist(webResource, event)
		if err != nil {
			logger.Err(err).Msg("error creating GDELT Event if it does not exist")
			return
		}
		if gdeltEvent == nil {
			return // skip creation
		}
		g.publishNewGDELTEvent(gdeltEvent)
	} else {
		webResource, err = g.createWebResourceAndGDELTEvent(event)
		if err != nil {
			logger.Err(err).Msg("error creating web resource and GDELT event")
			return
		}
		if webResource == nil {
			return // skip creation
		}
		g.publishNewWebResource(webResource)
		g.publishNewGDELTEvent(&webResource.GDELTEvent)
	}
}

func (g *Worker) cameoTopLevelEventCodeIsWhitelisted(event *events.Event) bool {
	whitelist := g.config.GDELTFetching.TopLevelCameoEventCodeWhitelist
	if len(whitelist) == 0 {
		return true
	}
	rootCode := event.EventRootCode

	for _, code := range whitelist {
		if code == rootCode {
			return true
		}
	}
	return false
}

func (g *Worker) createGDELTEventIfItDoesNotExist(
	webResource *models.WebResource,
	event *events.Event,
) (*models.GDELTEvent, error) {
	gdeltEventAssociation := g.db.Model(webResource).Association("GDELTEvent")
	if gdeltEventAssociation.Error != nil {
		return nil, fmt.Errorf("create WebResource.GDELTEvent association: %v", gdeltEventAssociation.Error)
	}

	if gdeltEventAssociation.Count() != 0 {
		g.logger.Debug().Uint("ID", webResource.ID).Msgf("a GDELTEvent already exists for this WebResource")
		return nil, nil
	}

	gdeltEvent, err := g.newGDELTEvent(event)
	if err != nil {
		return nil, fmt.Errorf("creating new GDELT Event model: %v", err)
	}
	gdeltEvent.WebResourceID = webResource.ID

	result := g.db.Create(&gdeltEvent)
	if result.Error != nil {
		return nil, fmt.Errorf("create GDELT Event for WebResource with URL %#v: %v",
			webResource.URL, result.Error)
	}

	return &gdeltEvent, nil
}

func (g *Worker) createWebResourceAndGDELTEvent(event *events.Event) (*models.WebResource, error) {
	gdeltEvent, err := g.newGDELTEvent(event)
	if err != nil {
		return nil, fmt.Errorf("creating new GDELT Event model: %v", err)
	}

	webResource := &models.WebResource{
		URL:        event.SourceURL,
		GDELTEvent: gdeltEvent,
	}
	result := *g.db.Create(webResource)
	if result.Error != nil {
		return nil, fmt.Errorf("create web resource and GDELT Event with URL %#v: %v",
			webResource.URL, result.Error)
	}
	return webResource, nil
}

func (g *Worker) newGDELTEvent(event *events.Event) (models.GDELTEvent, error) {
	dateAddedTime, err := event.DateAddedTime()
	if err != nil {
		return models.GDELTEvent{}, err
	}

	locationType := event.ActionGeo.Type.String()

	countryCode, err := event.ActionGeo.CountryCodeISO31661()
	if err != nil {
		return models.GDELTEvent{}, err
	}

	gdeltEvent := models.GDELTEvent{
		GlobalEventID:   uint(event.GlobalEventID),
		DateAdded:       dateAddedTime,
		LocationType:    sql.NullString{String: locationType, Valid: len(locationType) > 0},
		LocationName:    sql.NullString{String: event.ActionGeo.Fullname, Valid: len(event.ActionGeo.Fullname) > 0},
		CountryCode:     sql.NullString{String: countryCode, Valid: len(countryCode) > 0},
		Coordinates:     event.ActionGeo.PointCoordinates(),
		EventCategories: pgtype.VarcharArray{},
	}
	err = gdeltEvent.EventCategories.Set(event.AllCameoEventCodes())
	if err != nil {
		return models.GDELTEvent{}, err
	}
	return gdeltEvent, nil
}

func (g *Worker) publishNewWebResource(newWebResource *models.WebResource) {
	err := g.rmq.PublishID(g.config.GDELTFetching.NewWebResourceRoutingKey, newWebResource.ID)
	if err != nil {
		g.logger.Err(err).Uint("ID", newWebResource.ID).Msg("error publishing new web resource")
	}
}

func (g *Worker) publishNewGDELTEvent(newGDELTEvent *models.GDELTEvent) {
	err := g.rmq.PublishID(g.config.GDELTFetching.NewGDELTEventRoutingKey, newGDELTEvent.ID)
	if err != nil {
		g.logger.Err(err).Uint("ID", newGDELTEvent.ID).Msg("error publishing new GDELT Event")
	}
}
