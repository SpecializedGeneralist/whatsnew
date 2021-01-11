// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vectorizer

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcutils"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/vectorizer/labse"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"strings"
)

// Worker is a single worker that performs articles vectorization.
type Worker struct {
	config       configuration.Configuration
	db           *gorm.DB
	rmq          *rabbitmq.Client
	logger       zerolog.Logger
	labseGateway *labse.Gateway
}

// NewWorker creates a new Worker.
func NewWorker(
	config configuration.Configuration,
	db *gorm.DB,
	rmq *rabbitmq.Client,
	logger zerolog.Logger,
) (*Worker, error) {
	w := &Worker{
		config: config,
		db:     db,
		rmq:    rmq,
		logger: logger,
	}
	err := w.connectLaBSE()
	return w, err
}

func (w *Worker) Do(delivery amqp.Delivery) {
	w.logger.Debug().Msgf("processing msg %v", delivery.MessageId)

	webArticleID, err := rabbitmq.DecodeIDMessage(delivery.Body)
	if err != nil {
		w.logger.Err(err).Msg("error decoding ID message")
		w.sendNack(delivery)
		return
	}

	successfullyProcessed, err := w.processWebArticleID(webArticleID)
	if err != nil {
		w.logger.Err(err).Msg("error processing web article ID")
		w.sendNack(delivery)
		return
	}

	if successfullyProcessed {
		err = w.rmq.PublishID(w.config.Vectorizer.PubNewVectorizedWebArticleRoutingKey, webArticleID)
		if err != nil {
			w.logger.Err(err).Msgf("error publishing vectorized Web Article %d", webArticleID)
			w.sendNack(delivery)
			return
		}
	}

	w.sendAck(delivery)
}

func (w *Worker) processWebArticleID(webArticleID uint) (bool, error) {
	logger := w.logger.With().Uint("WebArticleID", webArticleID).Logger()

	webArticle, err := models.FindWebArticle(w.db, webArticleID)
	if err != nil {
		return false, err
	}

	title := strings.TrimSpace(webArticle.Title)
	if len(title) == 0 {
		logger.Debug().Msg("article with empty title skipped")
		return false, nil
	}

	vector, err := w.labseGateway.Vectorize(title)
	if err != nil {
		return false, fmt.Errorf("labse encode headline: %v", err)
	}

	data, err := tasks.Float32SliceToByteSlice(vector)
	if err != nil {
		return false, fmt.Errorf("error serializing encoding vector of Web Article %d: %v", webArticle.ID, err)
	}

	err = webArticle.Vector.Set(data)
	if err != nil {
		return false, fmt.Errorf("error setting vector of Web Article %d: %v", webArticle.ID, err)
	}

	result := w.db.Save(webArticle)
	if result.Error != nil {
		return false, fmt.Errorf("save vectorized Web Article: %v", result.Error)
	}

	return true, nil
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

func (w *Worker) connectLaBSE() error {
	w.logger.Info().Msgf("Creating: LaBSE gRPC connection [%s]", w.config.Vectorizer.LabseGrpcAddress)
	conn, err := grpcutils.OpenConnection(w.config.Vectorizer.LabseGrpcAddress, w.config.Vectorizer.LabseTLSDisable)
	if err != nil {
		w.logger.Fatal().Err(err).Msg("Failed to connect to LaBSE gRPC API")
	}
	w.labseGateway = labse.New(conn)
	return nil
}
