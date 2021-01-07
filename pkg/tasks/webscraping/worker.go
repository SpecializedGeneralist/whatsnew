// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webscraping

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/languagerecognition"
	goose "github.com/advancedlogic/GoOse"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// WebScrapingWorker is a single worker for scraping Web Resources.
type Worker struct {
	config  configuration.Configuration
	db      *gorm.DB
	rmq     *rabbitmq.Client
	logger  zerolog.Logger
	scraper goose.Goose
}

// NewWorker creates a new WebScrapingWorker.
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
		scraper: goose.New(),
	}
}

func (w *Worker) Do(delivery amqp.Delivery) {
	w.logger.Debug().Msgf("processing message %#v", delivery.MessageId)

	webResourceID, err := rabbitmq.DecodeIDMessage(delivery.Body)
	if err != nil {
		w.logger.Err(err).Msg("error decoding ID message")
		w.sendNack(delivery)
		return
	}

	webArticleID, err := w.processWebResourceID(webResourceID)
	if err != nil {
		w.logger.Err(err).Msgf("error processing web resource ID %d", webResourceID)
		w.sendNack(delivery)
		return
	}

	if webArticleID > 0 { // TODO: use custom error instead
		err = w.rmq.PublishID(w.config.WebScraping.PubNewWebArticleRoutingKey, webArticleID)
		if err != nil {
			w.logger.Err(err).Msgf("error publishing new Web Article %d", webArticleID)
			w.sendNack(delivery)
			return
		}
	}

	w.sendAck(delivery)
}

func (w *Worker) processWebResourceID(webResourceID uint) (uint, error) {
	webResource, err := w.getWebResource(webResourceID)
	if err != nil {
		return 0, err
	}

	logger := w.logger.With().
		Uint("WebResourceID", webResource.ID).Str("WebResourceURL", webResource.URL).Logger()

	if webResource.WebArticle.WebResourceID == webResource.ID {
		logger.Debug().Msgf("a Web Article already exists for this Web Resource")
		return 0, nil
	}

	body, err := w.fetchURLAndGetBody(webResource.URL)
	if err != nil {
		// TODO: recovery job needed
		logger.Warn().Err(err).Msg("error fetching remote web resource - article skipped")
		return 0, nil
	}

	if len(body) == 0 {
		logger.Debug().Msgf("empty body - article skipped")
		return 0, nil
	}

	article, err := w.scraper.ExtractFromRawHTML(body, webResource.URL)
	if err != nil {
		// TODO: consider recovery, but it might fail forever...
		logger.Warn().Err(err).Msg("error scraping HTML content - article skipped")
		return 0, nil
	}

	similarExists, err := w.webArticleWithSameTitleAlreadyExists(article.Title)
	if err != nil {
		return 0, err
	}
	if similarExists {
		logger.Debug().Str("title", article.Title).Msgf("web article with same title already exists - article skipped")
		return 0, nil
	}

	webArticle, err := w.creteWebArticle(article, webResource)
	if err != nil {
		return 0, fmt.Errorf("error creating new Web Article for Web Resource %#v: %v", webResource.URL, err)
	}

	if webArticle == nil {
		// TODO: skipping web article creation will create a problem when introducing recovery jobs.
		//       A solution might be to mark the WebResource as somehow "discarded"
		return 0, nil
	}

	return webArticle.ID, nil
}

func (w *Worker) getWebResource(webResourceID uint) (*models.WebResource, error) {
	webResource := &models.WebResource{}
	// TODO: use Joins?
	result := w.db.
		Preload("WebArticle").
		Preload("FeedItem").
		Preload("GDELTEvent").
		First(webResource, webResourceID)
	if result.Error != nil {
		return nil, fmt.Errorf("get WebResource: %v", result.Error)
	}
	return webResource, nil
}

func (w *Worker) webArticleWithSameTitleAlreadyExists(title string) (bool, error) {
	var count int64
	result := w.db.Model(&models.WebArticle{}).Where("title = ?", title).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("searching web articles with same title: %v", result.Error)
	}
	return count > 0, nil
}

func (w *Worker) creteWebArticle(
	article *goose.Article,
	webResource *models.WebResource,
) (*models.WebArticle, error) {
	// Only Feed Items already have a recognized language
	language := webResource.FeedItem.Language
	if len(language) == 0 {
		var hasLanguage bool
		language, hasLanguage = languagerecognition.RecognizeLanguage(article.Title)
		if !hasLanguage || !w.config.LanguageIsSupported(language) {
			w.logger.Debug().Str("language", language).
				Uint("WebResourceID", webResource.ID).
				Msg("recognized language is not supported")
			return nil, nil
		}
	}

	webArticle := &models.WebArticle{
		WebResourceID:      webResource.ID,
		Title:              article.Title,
		TitleUnmodified:    article.TitleUnmodified,
		CleanedText:        article.CleanedText,
		CanonicalLink:      article.CanonicalLink,
		TopImage:           article.TopImage,
		FinalURL:           article.FinalURL,
		ScrapedPublishDate: sql.NullTime{Valid: false},
		Language:           language,
		PublishDate: w.resolveArticleDate(
			webResource.FeedItem,
			webResource.GDELTEvent,
			article.PublishDate,
			webResource.CreatedAt,
		),
	}

	if article.PublishDate != nil {
		webArticle.ScrapedPublishDate.Time = *article.PublishDate
		webArticle.ScrapedPublishDate.Valid = true
	}

	result := w.db.Create(&webArticle)
	if result.Error != nil {
		return nil, fmt.Errorf("create WebArticle for WebResource %d: %v",
			webResource.ID, result.Error)
	}
	return webArticle, nil
}

func (w *Worker) resolveArticleDate(
	feedItem models.FeedItem,
	gdeltEvent models.GDELTEvent,
	scrapedPublishDate *time.Time,
	webResourceCreatedAt time.Time,
) time.Time {
	switch {
	case feedItem.ID != 0 && feedItem.PublishedAt.Valid:
		return feedItem.PublishedAt.Time.UTC()
	case gdeltEvent.ID != 0:
		return gdeltEvent.DateAdded.UTC()
	case scrapedPublishDate != nil:
		return scrapedPublishDate.UTC()
	default:
		return webResourceCreatedAt.UTC()
	}
}

func (w *Worker) sendNack(delivery amqp.Delivery) {
	err := delivery.Nack(false, true)
	if err != nil {
		w.logger.Err(err).Msg("error sending Nack")
	}
}

func (w *Worker) sendAck(delivery amqp.Delivery) {
	err := delivery.Ack(false)
	if err != nil {
		w.logger.Err(err).Msg("error sending Ack")
	}
}

func (w *Worker) fetchURLAndGetBody(url string) (strBody string, err error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrapf(err, "error creating new request for %#v", url)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_7) AppleWebKit/534.30 (KHTML, like Gecko) Chrome/12.0.742.91 Safari/534.30")

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "error performing request for %#v", url)
	}
	defer func() {
		if e := resp.Body.Close(); e != nil && err == nil {
			err = fmt.Errorf("close response body: %v", e)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrapf(err, "request for %#v returned status code %d", url, resp.StatusCode)
	}

	contentType := w.getContentType(resp)

	if !w.isTextOrHtmlContent(contentType) {
		return "", nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "error reading body from %#v", url)
	}

	enc := w.determineEncoding(body, contentType)
	transformReader := transform.NewReader(bytes.NewReader(body), enc.NewDecoder())
	transformedBody, err := ioutil.ReadAll(transformReader)
	if err != nil {
		return "", fmt.Errorf("transforming body to UTF-8: %v", err)
	}

	return strings.TrimSpace(string(transformedBody)), nil
}

func (w *Worker) getContentType(resp *http.Response) string {
	contentType, hasContentType := resp.Header["Content-Type"]
	if !hasContentType || len(contentType) == 0 {
		return ""
	}
	return contentType[0]
}

func (w *Worker) isTextOrHtmlContent(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text") || strings.Contains(contentType, "html")
}

func (w *Worker) determineEncoding(bytes []byte, contentType string) encoding.Encoding {
	e, _, _ := charset.DetermineEncoding(bytes, contentType)
	return e
}
