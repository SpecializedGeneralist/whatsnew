// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webscraper

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/database"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/languagerecognition"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	goose "github.com/advancedlogic/GoOse"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"net/http"
	"strings"
	"time"
)

// WebScraper implements a Faktory worker for scraping the Web, creating
// new WebArticles from existing WebResources.
type WebScraper struct {
	basemodelworker.Worker
	conf    config.WebScraper
	client  *http.Client
	scraper goose.Goose
}

// New creates a new WebScraper.
func New(conf config.WebScraper, db *gorm.DB, fk *faktory_worker.Manager) *WebScraper {
	ws := &WebScraper{
		conf: conf,
		client: &http.Client{
			Timeout: conf.RequestTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		scraper: goose.New(),
	}
	ws.Worker = basemodelworker.Worker{
		Name:        "WebScraper",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Perform:     ws.perform,
	}
	return ws
}

func (ws *WebScraper) perform(ctx context.Context, webResourceID uint) error {
	js := jobscheduler.New()

	err := ws.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// It's not possible to fetch and lock the WebResource while also
		// joining it with its related models with a single query.
		// We could lock the model and Preload the associations, but that would
		// result in many separate queries.
		// So we always perform exactly two queries. A first light query
		// simply locks the WebResource record, without getting any data.
		// Then we get the whole WebResource also joining all interesting
		// relations, with a single query and without having to worry about
		// the locking anymore.
		err := lockWebResource(tx, webResourceID)
		if err != nil {
			return err
		}
		wr, err := getWebResourceWithRelations(tx, webResourceID)
		if err != nil {
			return err
		}

		err = ws.processWebResource(ctx, tx, wr, js)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, ws.DB)
}

func lockWebResource(tx *gorm.DB, wrID uint) error {
	var wr *models.WebResource
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("ID").First(&wr, wrID)
	if res.Error != nil {
		return fmt.Errorf("error locking WebResource %d: %w", wrID, res.Error)
	}
	return nil
}

func getWebResourceWithRelations(tx *gorm.DB, wrID uint) (*models.WebResource, error) {
	var wr *models.WebResource

	// We ignore the Tweet relation on purpose, since tweets are always created
	// together with a WebArticle, so they don't need a separate scraping
	// process. In fact, this worker should never be scheduled for
	// tweet-resources in the first place.
	// The WebArticle, instead, is important to detect whether the scraping
	// was already performed.
	res := tx.
		Joins("WebArticle").
		Joins("FeedItem").
		Joins("GDELTEvent").
		First(&wr, wrID)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebResource %d: %w", wrID, res.Error)
	}
	return wr, nil
}

func (ws *WebScraper) processWebResource(ctx context.Context, tx *gorm.DB, wr *models.WebResource, js *jobscheduler.JobScheduler) error {
	logger := ws.Log.With().Uint("WebResource", wr.ID).Logger()

	if wr.WebArticle != nil {
		logger.Warn().Uint("WebArticle", wr.WebArticle.ID).Msg("a WebArticle already exists")
	}

	body, err := ws.scrapeURL(ctx, logger, wr.URL)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		logger.Debug().Msgf("empty body: skipping article")
		return nil
	}

	article, err := ws.extractFromHTML(body, wr.URL)
	if err != nil {
		logger.Warn().Err(err).Msgf("error extracting article from HTML: skipping article")
		return nil
	}

	similarExists, err := webArticleWithSameTitleExists(tx, article.Title)
	if err != nil {
		return err
	}
	if similarExists {
		logger.Debug().Str("Title", article.Title).Msgf("web article with same title already exists: article skipped")
		return nil
	}

	lang, langOk := resolveOrDetectLanguage(wr, article)
	if !langOk {
		logger.Warn().Str("Title", article.Title).Msg("failed to detect language")
		return nil
	}
	if !ws.languageIsAllowed(lang) {
		logger.Debug().Str("Title", article.Title).Str("Lang", lang).Msg("language is not allowed")
		return nil
	}

	webArticle := ws.newWebArticle(wr, article, lang)

	res := tx.Create(webArticle)
	if database.IsUniqueViolationError(res.Error) {
		logger.Warn().Err(res.Error).Msg("WebArticle creation constraint violation")
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("error creating WebArticle: %w", res.Error)
	}

	return js.AddJobs(ws.conf.NewWebArticleJobs, webArticle.ID)
}

func (ws *WebScraper) newWebArticle(wr *models.WebResource, article *goose.Article, lang string) *models.WebArticle {
	title := article.Title
	if wr.FeedItem != nil && len(wr.FeedItem.Title) > 0 {
		title = wr.FeedItem.Title
	}

	wa := &models.WebArticle{
		WebResourceID: wr.ID,
		Title:         title,
		TopImage: sql.NullString{
			String: article.TopImage,
			Valid:  len(article.TopImage) > 0,
		},
		Language:    lang,
		PublishDate: resolveArticleDate(wr, article),
	}

	if article.PublishDate != nil {
		wa.ScrapedPublishDate = sql.NullTime{
			Time:  *article.PublishDate,
			Valid: true,
		}
	}

	return wa
}

func (ws *WebScraper) languageIsAllowed(lang string) bool {
	for _, l := range ws.conf.LanguageFilter {
		if l == lang {
			return true
		}
	}
	return false
}

// extractFromHTML wraps a call to GoOse ExtractFromRawHTML, gracefully
// handling panics.
//
// For example, a panic was raised by a dependency package when parsing crappy
// HTML content, here:
// https://github.com/PuerkitoBio/goquery/blob/v1.6.0/manipulation.go#L579
func (ws *WebScraper) extractFromHTML(content, url string) (_ *goose.Article, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ExtractFromRawHTML panicked: %v", r)
		}
	}()
	return ws.scraper.ExtractFromRawHTML(content, url)
}

func (ws *WebScraper) scrapeURL(ctx context.Context, logger zerolog.Logger, url string) (_ string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request for %#v: %w", url, err)
	}
	req.Header.Add("User-Agent", ws.conf.UserAgent)

	resp, err := ws.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error performing request for %#v: %w", url, err)
	}
	defer func() {
		if e := resp.Body.Close(); e != nil && err == nil {
			err = fmt.Errorf("error closing response body: %w", e)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request for %#v returned status code %d", url, resp.StatusCode)
	}

	contentType := getContentType(resp)

	if !isTextOrHTMLContent(contentType) {
		logger.Debug().Msgf("ignoring content type %#v", contentType)
		return "", nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body for %#v: %w", url, err)
	}

	enc := determineEncoding(body, contentType)
	transformReader := transform.NewReader(bytes.NewReader(body), enc.NewDecoder())
	transformedBody, err := io.ReadAll(transformReader)
	if err != nil {
		return "", fmt.Errorf("error transforming body to UTF-8: %w", err)
	}

	return strings.TrimSpace(string(transformedBody)), nil
}

func resolveOrDetectLanguage(wr *models.WebResource, article *goose.Article) (string, bool) {
	// Only FeedItems already have a recognized language
	if wr.FeedItem != nil {
		return wr.FeedItem.Language, true
	}
	return languagerecognition.RecognizeLanguage(article.Title)
}

func resolveArticleDate(wr *models.WebResource, article *goose.Article) time.Time {
	switch {
	case wr.FeedItem != nil && wr.FeedItem.PublishedAt.Valid:
		return wr.FeedItem.PublishedAt.Time.UTC()
	case wr.GDELTEvent != nil:
		return wr.GDELTEvent.DateAdded.UTC()
	case article.PublishDate != nil:
		return article.PublishDate.UTC()
	default:
		return wr.CreatedAt.UTC()
	}
}

func webArticleWithSameTitleExists(tx *gorm.DB, title string) (bool, error) {
	var count int64
	result := tx.Model(&models.WebArticle{}).Where("title = ?", title).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("error counting web articles with same title: %w", result.Error)
	}
	return count > 0, nil
}

func getContentType(resp *http.Response) string {
	contentType, hasContentType := resp.Header["Content-Type"]
	if !hasContentType || len(contentType) == 0 {
		return ""
	}
	return contentType[0]
}

func isTextOrHTMLContent(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "text") || strings.Contains(ct, "html")
}

func determineEncoding(content []byte, contentType string) encoding.Encoding {
	enc, _, _ := charset.DetermineEncoding(content, contentType)
	return enc
}
