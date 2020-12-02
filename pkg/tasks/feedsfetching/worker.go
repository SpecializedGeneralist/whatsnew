// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feedsfetching

import (
	"database/sql"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/nlpodyssey/whatsnew/pkg/configuration"
	"github.com/nlpodyssey/whatsnew/pkg/models"
	"github.com/nlpodyssey/whatsnew/pkg/rabbitmq"
	"github.com/nlpodyssey/whatsnew/pkg/tasks/languagerecognition"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Worker struct {
	config     configuration.Configuration
	db         *gorm.DB
	rmq        *rabbitmq.Client
	logger     zerolog.Logger
	feedParser *gofeed.Parser
}

func NewWorker(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) *Worker {
	return &Worker{
		config:     config,
		db:         db,
		rmq:        rmq,
		logger:     logger,
		feedParser: gofeed.NewParser(),
	}
}

func (w *Worker) Do(feedID uint) {
	logger := w.logger.With().Uint("feedID", feedID).Logger()
	logger.Debug().Msg("processing feed ID")

	feed, err := models.FindFeed(w.db, feedID)
	if err != nil {
		logger.Err(err).Send()
		return
	}

	err = w.processFeed(logger, feed)
	if err != nil {
		logger.Warn().Err(err).Msg("error processing feed")
		return
	}

	err = w.updateSuccessfullyRetrievedFeed(feed)
	if err != nil {
		logger.Err(err).Msg("error updating retrieved feed")
	}
}

func (w *Worker) processFeed(logger zerolog.Logger, feed *models.Feed) error {
	parsedFeed, err := w.feedParser.ParseURL(feed.URL)
	if err != nil {
		return w.processFeedError(logger, feed, err)
	}

	for _, item := range parsedFeed.Items {
		w.processParsedFeedItem(logger, feed.ID, item)
	}
	return nil
}

func (w *Worker) processFeedError(logger zerolog.Logger, feed *models.Feed, err error) error {
	logger.Warn().Err(err).Msg("an error occurred when getting or parsing feed")

	feed.FailuresCount++
	logger.Warn().Msgf("incrementing failures count: %d", feed.FailuresCount)

	if feed.FailuresCount >= w.config.FeedsFetching.MaxAllowedFailures {
		logger.Warn().Msg("max allowed failures reached: soft-deleting feed")
		feed.DeletedAt = gorm.DeletedAt{Time: time.Now().UTC(), Valid: true}
	}

	result := w.db.Save(feed)
	if result.Error != nil {
		return fmt.Errorf("save feed with errors: %v", result.Error)
	}

	// Return original error to short-circuit FetchFeed operations
	return err
}

func (w *Worker) processParsedFeedItem(logger zerolog.Logger, feedID uint, item *gofeed.Item) {
	if w.feedItemIsTooOld(item) {
		logger.Debug().Time("PublishedParsed", *item.PublishedParsed).Msg("feed item publishing date is too old")
		return
	}

	language, hasLanguage := languagerecognition.RecognizeLanguage(item.Title)

	if !hasLanguage || !w.config.LanguageIsSupported(language) {
		logger.Debug().Str("language", language).Msg("recognized language is not supported")
		return
	}

	webResource, err := models.FindWebResourceByURL(w.db, item.Link)
	if err != nil {
		logger.Err(err).Msg("error finding web resource by URL")
		return
	}

	if webResource != nil {
		feedItem, err := w.createFeedItemIfItDoesNotExist(webResource, item, language, feedID)
		if err != nil {
			logger.Err(err).Msg("error creating feed item if it does not exist")
			return
		}
		if feedItem == nil {
			return // skip creation
		}
		w.publishNewFeedItem(logger, feedItem)
	} else {
		webResource, err = w.createWebResourceAndFeedItem(item, language, feedID)
		if err != nil {
			logger.Err(err).Msg("error creating web resource and feed item")
			return
		}
		if webResource == nil {
			return // skip creation
		}
		w.publishNewWebResource(logger, webResource)
		w.publishNewFeedItem(logger, &webResource.FeedItem)
	}
}

func (w *Worker) publishNewWebResource(logger zerolog.Logger, newWebResource *models.WebResource) {
	err := w.rmq.PublishID(w.config.FeedsFetching.NewWebResourceRoutingKey, newWebResource.ID)
	if err != nil {
		logger.Err(err).Uint("ID", newWebResource.ID).Msg("error publishing new web resource")
	}
}

func (w *Worker) publishNewFeedItem(logger zerolog.Logger, newFeedItem *models.FeedItem) {
	err := w.rmq.PublishID(w.config.FeedsFetching.NewFeedItemRoutingKey, newFeedItem.ID)
	if err != nil {
		logger.Err(err).Uint("ID", newFeedItem.ID).Msg("error publishing new feed item")
	}
}

func (w *Worker) createWebResourceAndFeedItem(
	item *gofeed.Item,
	language string,
	feedID uint,
) (*models.WebResource, error) {
	webResource := &models.WebResource{
		URL: item.Link,
	}
	result := w.db.Clauses(clause.OnConflict{DoNothing: true}).Create(webResource)
	if result.Error != nil {
		return nil, fmt.Errorf("create web resource with URL %#v: %v", webResource.URL, result.Error)
	}
	if result.RowsAffected == 0 {
		w.logger.Debug().Uint("feedID", feedID).Str("URL", webResource.URL).Msg("WebResource URL already exists")
		return nil, nil
	}

	webResource.FeedItem = w.newFeedItem(item, language, feedID)
	result = w.db.Save(webResource)
	if result.Error != nil {
		return nil, fmt.Errorf("create feed item for WebResource %d: %v", webResource.ID, result.Error)
	}

	return webResource, nil
}

func (w *Worker) createFeedItemIfItDoesNotExist(
	webResource *models.WebResource,
	item *gofeed.Item,
	language string,
	feedID uint,
) (*models.FeedItem, error) {
	feedItemAssociation := w.db.Model(webResource).Association("FeedItem")
	if feedItemAssociation.Error != nil {
		return nil, fmt.Errorf("create WebResource.FeedItem association: %v",
			feedItemAssociation.Error)
	}

	if feedItemAssociation.Count() != 0 {
		w.logger.Debug().Uint("ID", webResource.ID).Msgf("a FeedItem already exists for this WebResource")
		return nil, nil
	}

	feedItem := w.newFeedItem(item, language, feedID)
	feedItem.WebResourceID = webResource.ID

	result := w.db.Create(&feedItem)
	if result.Error != nil {
		return nil, fmt.Errorf("create feed item for WebResource with URL %#v: %v",
			webResource.URL, result.Error)
	}

	return &feedItem, nil
}

func (w *Worker) newFeedItem(item *gofeed.Item, language string, feedID uint) models.FeedItem {
	return models.FeedItem{
		FeedID:      feedID,
		Title:       item.Title,
		Description: makeNullString(item.Description),
		Content:     makeNullString(item.Content),
		Language:    language,
		PublishedAt: makeNullTime(item.PublishedParsed),
	}
}

func (w *Worker) feedItemIsTooOld(item *gofeed.Item) bool {
	return w.config.FeedsFetching.OmitFeedItemsPublishedBeforeEnabled &&
		item.PublishedParsed != nil &&
		item.PublishedParsed.Before(w.config.FeedsFetching.OmitFeedItemsPublishedBefore)
}

func (w *Worker) updateSuccessfullyRetrievedFeed(feed *models.Feed) error {
	feed.LastRetrievedAt.Time = time.Now().UTC()
	feed.LastRetrievedAt.Valid = true
	feed.FailuresCount = 0
	result := w.db.Save(feed)
	if result.Error != nil {
		return fmt.Errorf("update retrieved feed: %v", result.Error)
	}
	return nil
}

// TODO: find a better place for makeNullString and makeNullTime
func makeNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: len(s) > 0}
}
func makeNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
