// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feedfetcher

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/languagerecognition"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"time"
)

// FeedFetcher implements a Faktory worker for fetching feeds and creating new
// feed items.
type FeedFetcher struct {
	basemodelworker.Worker
	conf   config.FeedFetcher
	parser *gofeed.Parser
}

// New creates a new FeedFetcher.
func New(conf config.FeedFetcher, db *gorm.DB, fk *faktory_worker.Manager) *FeedFetcher {
	ff := &FeedFetcher{
		conf:   conf,
		parser: gofeed.NewParser(),
	}
	ff.Worker = basemodelworker.Worker{
		Name:        "FeedFetcher",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Queues:      conf.Queues,
		Perform:     ff.perform,
	}
	return ff
}

func (ff *FeedFetcher) perform(ctx context.Context, feedID uint) error {
	tx := ff.DB.WithContext(ctx)

	feed, err := getFeed(tx, feedID)
	if err != nil {
		return err
	}
	if !feed.Enabled {
		ff.Log.Warn().Msgf("skipping feed %d: not enabled", feed.ID)
		return nil
	}

	parsedFeed, parsedFeedError := ff.parseFeedURL(ctx, feed.URL)

	js := jobscheduler.New()
	err = tx.Transaction(func(tx *gorm.DB) error {
		err = ff.processFeed(tx, feed, js, parsedFeed, parsedFeedError)
		if err != nil {
			return err
		}
		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, ff.DB)
}

func getFeed(tx *gorm.DB, feedID uint) (*models.Feed, error) {
	var feed *models.Feed
	res := tx.First(&feed, feedID)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching Feed %d: %w", feedID, res.Error)
	}
	return feed, nil
}

func (ff *FeedFetcher) parseFeedURL(ctx context.Context, url string) (*gofeed.Feed, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ff.conf.RequestTimeout)
	defer cancel()
	return ff.parser.ParseURLWithContext(url, ctxTimeout)
}

func (ff *FeedFetcher) processFeed(
	tx *gorm.DB,
	feed *models.Feed,
	js *jobscheduler.JobScheduler,
	parsedFeed *gofeed.Feed,
	parsedFeedError error,
) error {
	if parsedFeedError != nil {
		ff.Log.Warn().Err(parsedFeedError).Msgf("error parsing feed %d", feed.ID)
		return ff.markFeedWithError(tx, feed, parsedFeedError)
	}

	err := ff.resetFeedErrors(tx, feed)
	if err != nil {
		return err
	}

	for _, item := range parsedFeed.Items {
		err = ff.processParsedFeedItem(tx, feed, item, js)
		if err != nil {
			return fmt.Errorf("error processing parsed feed item: %w", err)
		}
	}

	feed.LastRetrievedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	err = models.OptimisticSave(tx, feed)
	if err != nil {
		return fmt.Errorf("error updating Feed.LastRetrievedAt: %w", err)
	}

	return nil
}

func (ff *FeedFetcher) processParsedFeedItem(
	tx *gorm.DB,
	feed *models.Feed,
	item *gofeed.Item,
	js *jobscheduler.JobScheduler,
) error {
	logger := ff.Log.With().Uint("Feed", feed.ID).Str("Link", item.Link).Logger()

	if ff.itemIsTooOld(item) {
		logger.Debug().Time("PublishedParsed", *item.PublishedParsed).Msg("item is too old")
		return nil
	}

	lang, langOk := languagerecognition.RecognizeLanguage(item.Title)
	if !langOk {
		logger.Warn().Str("Title", item.Title).Msg("failed to detect language")
		return nil
	}
	if !ff.languageIsAllowed(lang) {
		logger.Debug().Str("Title", item.Title).Str("Lang", lang).Msg("language is not allowed")
		return nil
	}

	webResource, err := findWebResource(tx, item.Link)
	if err != nil {
		return err
	}

	feedItem := &models.FeedItem{
		FeedID:      feed.ID,
		Title:       item.Title,
		Description: item.Description,
		Content:     item.Content,
		Language:    lang,
		PublishedAt: makeNullTime(item.PublishedParsed),
	}

	if webResource != nil {
		logger = logger.With().Uint("WebResource", webResource.ID).Logger()

		if webResource.FeedItem != nil {
			logger.Debug().Uint("FeedItem", webResource.FeedItem.ID).Msg("a FeedItem already exists")
			return nil
		}

		feedItem.WebResourceID = webResource.ID
		return createFeedItem(tx, logger, feedItem)
	}

	webResource = &models.WebResource{
		URL:      item.Link,
		FeedItem: feedItem,
	}

	res := tx.Create(webResource)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("WebResource and FeedItem creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating WebResource: %w", res.Error)
	}
	return js.AddJobs(ff.conf.NewWebResourceJobs, webResource.ID)
}

func createFeedItem(tx *gorm.DB, logger zerolog.Logger, fi *models.FeedItem) error {
	res := tx.Create(fi)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("FeedItem creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating FeedItem: %w", res.Error)
	}
	return nil
}

func (ff *FeedFetcher) markFeedWithError(tx *gorm.DB, feed *models.Feed, feedError error) error {
	feed.LastError = sql.NullString{Valid: true, String: feedError.Error()}
	feed.FailuresCount++
	feed.Enabled = feed.FailuresCount <= ff.conf.MaxAllowedFailures

	err := models.OptimisticSave(tx, feed)
	if err != nil {
		return fmt.Errorf("error saving feed (marked with error): %w", err)
	}
	return nil
}

func (ff *FeedFetcher) resetFeedErrors(tx *gorm.DB, feed *models.Feed) error {
	// Don't waste an UPDATE if there's nothing to change.
	if !feed.LastError.Valid && feed.FailuresCount == 0 {
		return nil
	}

	feed.LastError = sql.NullString{Valid: false, String: ""}
	feed.FailuresCount = 0

	err := models.OptimisticSave(tx, feed)
	if err != nil {
		return fmt.Errorf("error saving feed (resetting errors): %w", err)
	}
	return nil
}

func findWebResource(tx *gorm.DB, url string) (*models.WebResource, error) {
	var webResource *models.WebResource
	result := tx.Joins("FeedItem").Limit(1).Find(&webResource, "url = ?", url)
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching WebResource by URL %#v: %w", url, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return webResource, nil
}

func (ff *FeedFetcher) itemIsTooOld(item *gofeed.Item) bool {
	return ff.conf.OmitItemsPublishedBefore.Enabled &&
		item.PublishedParsed != nil &&
		item.PublishedParsed.Before(ff.conf.OmitItemsPublishedBefore.Time)
}

func (ff *FeedFetcher) languageIsAllowed(lang string) bool {
	for _, l := range ff.conf.LanguageFilter {
		if l == lang {
			return true
		}
	}
	return false
}

func makeNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
