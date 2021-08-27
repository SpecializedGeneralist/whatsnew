// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vectorizer

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/hnswclient"
	"github.com/SpecializedGeneralist/whatsnew/pkg/jobscheduler"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/workers/basemodelworker"
	"github.com/contribsys/faktory_worker_go"
	"github.com/nlpodyssey/spago/pkg/mat32"
	bert_grpcapi "github.com/nlpodyssey/spago/pkg/nlp/transformers/bert/grpcapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

// Vectorizer implements a Faktory worker for creating and storing a vector
// representation of WebArticles' titles.
type Vectorizer struct {
	basemodelworker.Worker
	conf       config.Vectorizer
	bertClient bert_grpcapi.BERTClient
	hnswClient *hnswclient.Client
}

// New creates a new WebScraper.
func New(conf config.Vectorizer, db *gorm.DB, bertConn *grpc.ClientConn, hnswClient *hnswclient.Client, fk *faktory_worker.Manager) *Vectorizer {
	v := &Vectorizer{
		conf:       conf,
		bertClient: bert_grpcapi.NewBERTClient(bertConn),
		hnswClient: hnswClient,
	}
	v.Worker = basemodelworker.Worker{
		Name:        "Vectorizer",
		DB:          db,
		FK:          fk,
		Log:         log.Logger.Level(zerolog.Level(conf.LogLevel)),
		Concurrency: conf.Concurrency,
		Perform:     v.perform,
	}
	return v
}

func (v *Vectorizer) perform(ctx context.Context, webArticleID uint) error {
	js := jobscheduler.New()

	err := v.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		wa, err := getLockedWebArticle(tx, webArticleID)
		if err != nil {
			return err
		}

		err = v.processWebArticle(ctx, tx, wa, js)
		if err != nil {
			return err
		}

		return js.CreatePendingJobs(tx)
	})
	if err != nil {
		return err
	}

	return js.PushJobsAndDeletePendingJobs(ctx, v.DB)
}

func getLockedWebArticle(tx *gorm.DB, id uint) (*models.WebArticle, error) {
	var wa *models.WebArticle
	res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wa, id)
	if res.Error != nil {
		return nil, fmt.Errorf("error fetching WebArticle %d: %w", id, res.Error)
	}
	return wa, nil
}

func (v *Vectorizer) processWebArticle(ctx context.Context, tx *gorm.DB, wa *models.WebArticle, js *jobscheduler.JobScheduler) error {
	logger := v.Log.With().Uint("WebArticle", wa.ID).Logger()

	if wa.HasVector {
		logger.Warn().Msg("this WebArticle already has a vector")
		return nil
	}

	title := strings.TrimSpace(wa.Title)
	if len(title) == 0 {
		logger.Debug().Msg("empty title - web article skipped")
		return nil
	}

	vector, err := v.vectorize(ctx, title)
	if err != nil {
		return err
	}

	err = v.hnswClient.Insert(ctx, wa.ID, wa.PublishDate, vector)
	if err != nil {
		return err
	}

	wa.HasVector = true
	res := tx.Save(wa)
	if res.Error != nil {
		return fmt.Errorf("error saving WebArticle: %w", res.Error)
	}

	return js.AddJobs(v.conf.VectorizedWebArticleJobs, wa.ID)
}

// vectorize returns a dense vector representation of the given text.
// It is expected to work well with models such as LaBSE (Language-agnostic
// BERT Sentence Embedding).
//
// It simply calls the remote BERT Encode method to get a vector, which is
// then normalized and returned.
func (v *Vectorizer) vectorize(ctx context.Context, text string) ([]float32, error) {
	request := &bert_grpcapi.EncodeRequest{Text: text}
	encoding, err := v.bertClient.Encode(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("BERT encoding error: %w", err)
	}
	return normalize(encoding.Vector), nil
}

func normalize(xs []float32) []float32 {
	return mat32.NewVecDense(xs).Normalize2().Data()
}
