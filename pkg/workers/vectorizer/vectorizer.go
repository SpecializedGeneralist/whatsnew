// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vectorizer

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcconn"
	"github.com/SpecializedGeneralist/whatsnew/pkg/hnswclient"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/jackc/pgtype"
	"github.com/nlpodyssey/spago/pkg/mat32"
	bertgrpcapi "github.com/nlpodyssey/spago/pkg/nlp/transformers/bert/grpcapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"strings"
)

// Vectorizer implements a Faktory worker for creating and storing a vector
// representation of WebArticles' titles.
type Vectorizer struct {
	basemodelworker.Worker
	conf     config.Vectorizer
	hnswConf config.HNSW
	bertgrpcapi.BERTClient
}

// New creates a new WebScraper.
func New(
	conf config.Vectorizer,
	hnswConf config.HNSW,
	db *gorm.DB,
	fk *faktory_worker.Manager,
) *Vectorizer {
	v := &Vectorizer{
		conf:     conf,
		hnswConf: hnswConf,
	}
	v.Worker = basemodelworker.Worker{
		Name:        "Vectorizer",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Queues:      conf.Queues,
		Perform:     v.perform,
	}
	return v
}

func (v *Vectorizer) perform(ctx context.Context, webArticleID uint) error {
	tx := v.DB.WithContext(ctx)

	wa, err := getWebArticle(tx, webArticleID)
	if err != nil {
		return err
	}

	vecModel, err := v.processWebArticle(ctx, wa)
	if err != nil {
		return err
	}
	if vecModel == nil {
		return nil // skipped
	}

	js := jobscheduler.New()
	err = tx.Transaction(func(tx *gorm.DB) error {
		res := tx.Create(vecModel)
		if res.Error != nil {
			return fmt.Errorf("error saving Vector: %w", res.Error)
		}

		return js.AddJobsAndCreatePendingJobs(tx, v.conf.VectorizedWebArticleJobs, wa.ID)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, v.DB)
}

func getWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Preload("Vector").First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (v *Vectorizer) processWebArticle(
	ctx context.Context,
	wa *models.WebArticle,
) (*models.Vector, error) {
	logger := v.Log.With().Uint("WebArticle", wa.ID).Logger()

	if wa.Vector != nil {
		logger.Warn().Msg("this WebArticle already has a vector")
		return nil, nil
	}

	title := strings.TrimSpace(wa.Title)
	if wa.TranslatedTitle.Valid {
		title = strings.TrimSpace(wa.TranslatedTitle.String)
	}

	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return nil, nil
	}

	hnswConn, err := grpcconn.Dial(ctx, v.hnswConf.Server)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := hnswConn.Close(); err != nil {
			v.Log.Err(err).Msg("error closing HNSW connection")
		}
	}()
	hnswClient := hnswclient.New(hnswConn, v.hnswConf.Index)

	vector, err := v.vectorize(ctx, title)
	if err != nil {
		return nil, err
	}

	err = hnswClient.Insert(ctx, wa.ID, wa.PublishDate, vector)
	if err != nil {
		return nil, err
	}

	vectorModel := &models.Vector{
		WebArticleID: wa.ID,
		Data:         new(pgtype.Float4Array),
	}
	err = vectorModel.Data.Set(vector)
	if err != nil {
		return nil, fmt.Errorf("error setting Vector data: %w", err)
	}

	return vectorModel, nil
}

// vectorize returns a dense vector representation of the given text.
// It is expected to work well with models such as LaBSE (Language-agnostic
// BERT Sentence Embedding).
//
// It simply calls the remote BERT Encode method to get a vector, which is
// then normalized and returned.
func (v *Vectorizer) vectorize(ctx context.Context, text string) ([]float32, error) {
	bertConn, err := grpcconn.Dial(ctx, v.conf.SpagoBERTServer)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := bertConn.Close(); err != nil {
			v.Log.Err(err).Msg("error closing BERT connection")
		}
	}()
	bertClient := bertgrpcapi.NewBERTClient(bertConn)

	request := &bertgrpcapi.EncodeRequest{Text: text}
	encoding, err := bertClient.Encode(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("BERT encoding error: %w", err)
	}
	return normalize(encoding.Vector), nil
}

func normalize(xs []float32) []float32 {
	return mat32.NewVecDense(xs).Normalize2().Data()
}
