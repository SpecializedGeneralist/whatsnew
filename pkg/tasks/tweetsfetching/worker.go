// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tweetsfetching

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/languagerecognition"
	"github.com/jackc/pgtype"
	"github.com/n0madic/twitter-scraper"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Worker struct {
	config  configuration.Configuration
	db      *gorm.DB
	rmq     *rabbitmq.Client
	logger  zerolog.Logger
	scraper *twitterscraper.Scraper
}

const maxTweetsNumber = 50

func NewWorker(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) *Worker {
	return &Worker{
		config:  config,
		db:      db,
		rmq:     rmq,
		logger:  logger,
		scraper: twitterscraper.New(),
	}
}

func (w *Worker) Do(twitterSourceID uint) {
	logger := w.logger.With().Uint("twitterSourceID", twitterSourceID).Logger()
	logger.Debug().Msg("processing twitter-source ID")

	ts, err := models.FindTwitterSource(w.db, twitterSourceID)
	if err != nil {
		logger.Err(err).Send()
		return
	}

	err = w.processTwitterSource(logger, ts)
	if err != nil {
		logger.Warn().Err(err).Msg("error processing twitter-source")
		return
	}

	err = w.updateSuccessfullyRetrievedTwitterSource(ts)
	if err != nil {
		logger.Err(err).Msg("error updating retrieved twitter-source")
	}
}

func (w *Worker) processTwitterSource(logger zerolog.Logger, ts *models.TwitterSource) error {
	var ch <-chan *twitterscraper.Result

	switch ts.Type {
	case models.UserTwitterSource:
		ch = w.scraper.GetTweets(context.Background(), ts.Value, maxTweetsNumber)
	case models.SearchTwitterSource:
		ch = w.scraper.SearchTweets(context.Background(), ts.Value, maxTweetsNumber)
	default:
		return fmt.Errorf("unexpected twitter-source type %#v", ts.Type)
	}

	for result := range ch {
		if result.Error != nil {
			w.logger.Err(result.Error).Uint("TS-ID", ts.ID).Msg("twitter result error")
			continue
		}
		w.processTweet(logger, ts.ID, &result.Tweet)
	}

	return nil
}

func (w *Worker) processTweet(logger zerolog.Logger, twitterSourceID uint, scrapedTweet *twitterscraper.Tweet) {
	if w.tweetIsTooOld(scrapedTweet) {
		logger.Debug().Time("TimeParsed", scrapedTweet.TimeParsed).Msg("the tweet is too old")
		return
	}

	language, hasLanguage := languagerecognition.RecognizeLanguage(scrapedTweet.Text)

	if !hasLanguage || !w.config.LanguageIsSupported(language) {
		logger.Debug().Str("language", language).Msg("recognized language is not supported")
		return
	}

	webResource, err := models.FindWebResourceByURL(w.db, scrapedTweet.PermanentURL)
	if err != nil {
		logger.Err(err).Msg("error finding web resource by URL")
		return
	}

	if webResource != nil {
		tweet, err := w.createTweetIfItDoesNotExist(webResource, scrapedTweet, twitterSourceID)
		if err != nil {
			logger.Err(err).Msg("error creating tweet if it does not exist")
			return
		}
		if tweet != nil {
			w.publishNewTweet(logger, tweet)
		}

		webArticle, err := w.createWebArticleIfItDoesNotExist(webResource, scrapedTweet, language, twitterSourceID)
		if err != nil {
			logger.Err(err).Msg("error creating web-article if it does not exist")
			return
		}
		if webArticle != nil {
			w.publishNewWebArticle(logger, webArticle)
		}
	} else {
		webResource, err = w.createEntireWebResource(scrapedTweet, language, twitterSourceID)
		if err != nil {
			logger.Err(err).Msg("error creating web resource and tweet and web article")
			return
		}
		if webResource == nil {
			return // skip creation
		}
		w.publishNewWebResource(logger, webResource)
		w.publishNewTweet(logger, &webResource.Tweet)
		w.publishNewWebArticle(logger, &webResource.WebArticle)
	}
}

func (w *Worker) publishNewTweet(logger zerolog.Logger, newTweet *models.Tweet) {
	routingKey := w.config.TweetsFetching.NewTweetRoutingKey
	if len(routingKey) == 0 {
		return
	}
	err := w.rmq.PublishID(routingKey, newTweet.ID)
	if err != nil {
		logger.Err(err).Uint("ID", newTweet.ID).Msg("error publishing new tweet")
	}
}

func (w *Worker) publishNewWebArticle(logger zerolog.Logger, newWebArticle *models.WebArticle) {
	routingKey := w.config.TweetsFetching.NewWebArticleRoutingKey
	if len(routingKey) == 0 {
		return
	}
	err := w.rmq.PublishID(routingKey, newWebArticle.ID)
	if err != nil {
		logger.Err(err).Uint("ID", newWebArticle.ID).Msg("error publishing new web-article")
	}
}

func (w *Worker) publishNewWebResource(logger zerolog.Logger, newWebResource *models.WebResource) {
	routingKey := w.config.TweetsFetching.NewWebResourceRoutingKey
	if len(routingKey) == 0 {
		return
	}
	err := w.rmq.PublishID(routingKey, newWebResource.ID)
	if err != nil {
		logger.Err(err).Uint("ID", newWebResource.ID).Msg("error publishing new web-resource")
	}
}

// createEntireWebResource creates a new WebResource, Tweet, and WebArticle.
func (w *Worker) createEntireWebResource(
	scrapedTweet *twitterscraper.Tweet,
	language string,
	twitterSourceID uint,
) (*models.WebResource, error) {
	webResource := &models.WebResource{
		URL: scrapedTweet.PermanentURL,
	}
	result := w.db.Clauses(clause.OnConflict{DoNothing: true}).Create(webResource)
	if result.Error != nil {
		return nil, fmt.Errorf("create web resource with URL %#v: %v", webResource.URL, result.Error)
	}
	if result.RowsAffected == 0 {
		w.logger.Debug().Uint("twitterSourceID", twitterSourceID).Str("URL", webResource.URL).Msg("WebResource URL already exists")
		return nil, nil
	}

	webResource.Tweet = w.newTweet(scrapedTweet, twitterSourceID)
	webResource.WebArticle = w.newWebArticle(scrapedTweet, language, twitterSourceID)

	result = w.db.Save(webResource)
	if result.Error != nil {
		return nil, fmt.Errorf("create Tweet and WebArticle for WebResource %d: %v", webResource.ID, result.Error)
	}

	return webResource, nil
}

func (w *Worker) createTweetIfItDoesNotExist(
	webResource *models.WebResource,
	scrapedTweet *twitterscraper.Tweet,
	twitterSourceID uint,
) (*models.Tweet, error) {
	tweetAssociation := w.db.Model(webResource).Association("Tweet")
	if tweetAssociation.Error != nil {
		return nil, fmt.Errorf("create WebResource.Tweet association: %v", tweetAssociation.Error)
	}

	if tweetAssociation.Count() != 0 {
		w.logger.Debug().Uint("ID", webResource.ID).Msgf("a Tweet already exists for this WebResource")
		return nil, nil
	}

	tweet := w.newTweet(scrapedTweet, twitterSourceID)
	tweet.WebResourceID = webResource.ID

	result := w.db.Create(&tweet)
	if result.Error != nil {
		return nil, fmt.Errorf("create tweet for WebResource with URL %#v: %v", webResource.URL, result.Error)
	}

	return &tweet, nil
}

func (w *Worker) createWebArticleIfItDoesNotExist(
	webResource *models.WebResource,
	scrapedTweet *twitterscraper.Tweet,
	language string,
	twitterSourceID uint,
) (*models.WebArticle, error) {
	webArticleAssoc := w.db.Model(webResource).Association("WebArticle")
	if webArticleAssoc.Error != nil {
		return nil, fmt.Errorf("create WebResource.WebArticle association: %v", webArticleAssoc.Error)
	}

	if webArticleAssoc.Count() != 0 {
		w.logger.Debug().Uint("ID", webResource.ID).Msgf("a WebArticle already exists for this WebResource")
		return nil, nil
	}

	webArticle := w.newWebArticle(scrapedTweet, language, twitterSourceID)
	webArticle.WebResourceID = webResource.ID

	result := w.db.Create(&webArticle)
	if result.Error != nil {
		return nil, fmt.Errorf("create web-article for WebResource with URL %#v: %v", webResource.URL, result.Error)
	}

	return &webArticle, nil
}

func (w *Worker) newTweet(scrapedTweet *twitterscraper.Tweet, twitterSourceID uint) models.Tweet {
	return models.Tweet{
		TwitterSourceID: twitterSourceID,
		UpstreamID:      scrapedTweet.ID,
		Text:            scrapedTweet.Text,
		PublishedAt:     scrapedTweet.TimeParsed,
		Username:        scrapedTweet.Username,
		UserID:          scrapedTweet.UserID,
	}
}

func (w *Worker) newWebArticle(scrapedTweet *twitterscraper.Tweet, language string, twitterSourceID uint) models.WebArticle {
	topImage := ""
	if len(scrapedTweet.Photos) > 0 {
		topImage = scrapedTweet.Photos[0]
	}
	return models.WebArticle{
		Title:           scrapedTweet.Text,
		TitleUnmodified: "",
		CleanedText:     "",
		CanonicalLink:   "",
		TopImage:        topImage,
		FinalURL:        "",
		ScrapedPublishDate: sql.NullTime{
			Time:  scrapedTweet.TimeParsed,
			Valid: true,
		},
		Language:    language,
		PublishDate: scrapedTweet.TimeParsed,
		Vector:      pgtype.Bytea{Bytes: nil, Status: pgtype.Null},
	}
}

func (w *Worker) tweetIsTooOld(scrapedTweet *twitterscraper.Tweet) bool {
	return w.config.TweetsFetching.OmitTweetsPublishedBeforeEnabled &&
		scrapedTweet.TimeParsed.Before(w.config.TweetsFetching.OmitTweetsPublishedBefore)
}

func (w *Worker) updateSuccessfullyRetrievedTwitterSource(ts *models.TwitterSource) error {
	ts.LastRetrievedAt.Time = time.Now().UTC()
	ts.LastRetrievedAt.Valid = true
	result := w.db.Save(ts)
	if result.Error != nil {
		return fmt.Errorf("update retrieved twitter-source: %v", result.Error)
	}
	return nil
}
