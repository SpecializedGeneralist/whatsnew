// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package twitterscraper

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
	"github.com/n0madic/twitter-scraper"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"time"
)

// TwitterScraper implements a Faktory worker for scraping Twitter according
// to a specific source and creating new web articles.
type TwitterScraper struct {
	basemodelworker.Worker
	conf    config.TwitterScraper
	scraper *twitterscraper.Scraper
}

// New creates a new TwitterScraper.
func New(conf config.TwitterScraper, db *gorm.DB, fk *faktory_worker.Manager) *TwitterScraper {
	ts := &TwitterScraper{
		conf:    conf,
		scraper: twitterscraper.New(),
	}
	ts.Worker = basemodelworker.Worker{
		Name:        "TwitterScraper",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Queues:      conf.Queues,
		Perform:     ts.perform,
	}
	return ts
}

func (ts *TwitterScraper) perform(ctx context.Context, twitterSourceID uint) error {
	tx := ts.DB.WithContext(ctx)

	src, err := getTwitterSource(tx, twitterSourceID)
	if err != nil {
		return err
	}
	if !src.Enabled {
		ts.Log.Warn().Msgf("skipping TwitterSource %d: not enabled", src.ID)
		return nil
	}

	js := jobscheduler.New()
	err = tx.Transaction(func(tx *gorm.DB) error {
		err = ts.processTwitterSource(ctx, tx, src, js)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, ts.DB)
}

func getTwitterSource(tx *gorm.DB, tsID uint) (*models.TwitterSource, error) {
	var src *models.TwitterSource
	res := tx.First(&src, tsID)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching TwitterSource %d: %w", tsID, res.Error)
	}
	return src, nil
}

func (ts *TwitterScraper) processTwitterSource(
	ctx context.Context,
	tx *gorm.DB,
	src *models.TwitterSource,
	js *jobscheduler.JobScheduler,
) error {
	var ch <-chan *twitterscraper.TweetResult

	switch src.Type {
	case models.UserTwitterSource:
		ch = ts.scraper.GetTweets(ctx, src.Text, ts.conf.MaxTweetsNumber)
	case models.SearchTwitterSource:
		ch = ts.scraper.SearchTweets(ctx, src.Text, ts.conf.MaxTweetsNumber)
	default:
		err := fmt.Errorf("unexpected twitter-source type %#v", src.Type)
		ts.Log.Err(err).Msgf("error reading TwitterSource %d", src.ID)
		return ts.markSourceWithError(tx, src, err)
	}

	for tr := range ch {
		if tr.Error != nil {
			ts.Log.Err(tr.Error).Msgf("error reading TwitterSource %d results", src.ID)
			return ts.markSourceWithError(tx, src, tr.Error)
		}
		err := ts.processTweet(tx, src, tr.Tweet, js)
		if err != nil {
			return fmt.Errorf("error processing tweet result: %w", err)
		}
	}

	return ts.resetSourceErrors(tx, src)
}

func (ts *TwitterScraper) processTweet(
	tx *gorm.DB,
	src *models.TwitterSource,
	scrapedTweet twitterscraper.Tweet,
	js *jobscheduler.JobScheduler,
) error {
	logger := ts.Log.With().Uint("TwitterSource", src.ID).Str("ScrapedTweet", scrapedTweet.ID).Logger()

	if ts.tweetIsTooOld(scrapedTweet) {
		logger.Debug().Time("TimeParsed", scrapedTweet.TimeParsed).Msg("the tweet is too old")
		return nil
	}

	lang, langOk := languagerecognition.RecognizeLanguage(scrapedTweet.Text)
	if !langOk {
		logger.Warn().Str("Text", scrapedTweet.Text).Msg("failed to detect language")
		return nil
	}
	if !ts.languageIsAllowed(lang) {
		logger.Debug().Str("Text", scrapedTweet.Text).Str("Lang", lang).Msg("language is not allowed")
		return nil
	}

	webResource, err := findWebResource(tx, scrapedTweet.PermanentURL)
	if err != nil {
		return err
	}

	tweet := newTweet(src, scrapedTweet)
	webArticle := newWebArticle(scrapedTweet, lang)

	if webResource != nil {
		logger = logger.With().Uint("WebResource", webResource.ID).Logger()

		if webResource.Tweet != nil {
			logger.Debug().Uint("Tweet", webResource.Tweet.ID).Msg("a Tweet already exists")
		} else {
			tweet.WebResourceID = webResource.ID
			err = createTweet(tx, logger, tweet)
			if err != nil {
				return err
			}
		}

		if webResource.WebArticle != nil {
			logger.Debug().Uint("WebArticle", webResource.WebArticle.ID).Msg("a WebArticle already exists")
			return nil
		}

		webArticle.WebResourceID = webResource.ID
		err = createWebArticle(tx, logger, webArticle)
		if err != nil {
			return err
		}
		return js.AddJobs(ts.conf.NewWebArticleJobs, webArticle.ID)
	}

	webResource = &models.WebResource{
		URL:        scrapedTweet.PermanentURL,
		Tweet:      tweet,
		WebArticle: webArticle,
	}

	res := tx.Create(webResource)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("WebResource, Tweet and WebArticle creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating WebResource: %w", res.Error)
	}
	return js.AddJobs(ts.conf.NewWebArticleJobs, webResource.WebArticle.ID)
}

func newTweet(src *models.TwitterSource, scrapedTweet twitterscraper.Tweet) *models.Tweet {
	return &models.Tweet{
		TwitterSourceID: src.ID,
		UpstreamID:      scrapedTweet.ID,
		Text:            scrapedTweet.Text,
		PublishedAt:     scrapedTweet.TimeParsed,
		Username:        scrapedTweet.Username,
		UserID:          scrapedTweet.UserID,
	}
}

func newWebArticle(scrapedTweet twitterscraper.Tweet, language string) *models.WebArticle {
	wa := &models.WebArticle{
		Title: scrapedTweet.Text,
		ScrapedPublishDate: sql.NullTime{
			Valid: true,
			Time:  scrapedTweet.TimeParsed,
		},
		Language:    language,
		PublishDate: scrapedTweet.TimeParsed,
	}
	if len(scrapedTweet.Photos) > 0 {
		wa.TopImage = sql.NullString{
			Valid:  true,
			String: scrapedTweet.Photos[0],
		}
	}
	return wa
}

func createTweet(tx *gorm.DB, logger zerolog.Logger, tw *models.Tweet) error {
	res := tx.Create(tw)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("Tweet creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating Tweet: %w", res.Error)
	}
	return nil
}

func createWebArticle(tx *gorm.DB, logger zerolog.Logger, wa *models.WebArticle) error {
	res := tx.Create(wa)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("WebArticle creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating WebArticle: %w", res.Error)
	}
	return nil
}

func (ts *TwitterScraper) markSourceWithError(tx *gorm.DB, src *models.TwitterSource, sourceErr error) error {
	src.LastError = sql.NullString{Valid: true, String: sourceErr.Error()}
	src.FailuresCount++

	err := models.OptimisticSave(tx, src)
	if err != nil {
		return fmt.Errorf("error saving TwitterSource (marked with error): %w", err)
	}
	return nil
}

func (ts *TwitterScraper) resetSourceErrors(tx *gorm.DB, src *models.TwitterSource) error {
	src.LastRetrievedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	src.LastError = sql.NullString{Valid: false, String: ""}
	src.FailuresCount = 0

	err := models.OptimisticSave(tx, src)
	if err != nil {
		return fmt.Errorf("error saving TwitterSource (resetting errors): %w", err)
	}
	return nil
}

func findWebResource(tx *gorm.DB, url string) (*models.WebResource, error) {
	var webResource *models.WebResource
	result := tx.Joins("Tweet").Joins("WebArticle").Limit(1).Find(&webResource, "url = ?", url)
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching WebResource by URL %#v: %w", url, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return webResource, nil
}

func (ts *TwitterScraper) tweetIsTooOld(tweet twitterscraper.Tweet) bool {
	return ts.conf.OmitTweetsPublishedBefore.Enabled &&
		tweet.TimeParsed.Before(ts.conf.OmitTweetsPublishedBefore.Time)
}

func (ts *TwitterScraper) languageIsAllowed(lang string) bool {
	for _, l := range ts.conf.LanguageFilter {
		if l == lang {
			return true
		}
	}
	return false
}
