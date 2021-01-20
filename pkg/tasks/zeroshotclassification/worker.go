// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zeroshotclassification

import (
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/grpcutils"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
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
	spagoGateway *Gateway
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
	err := w.connectGRPC()
	return w, err
}

// Do performs the job of a zero-shot classification worker.
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
		err = w.rmq.PublishID(w.config.ZeroShotClassification.PubRoutingKey, webArticleID)
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

	conf := w.config.ZeroShotClassification
	zsc, err := w.spagoGateway.ClassifyNLI(title, conf.HypothesisTemplate, conf.PossibleLabels, conf.MultiClass)
	if err != nil {
		return false, fmt.Errorf("ClassifyNLI error: %w", err)
	}

	if webArticle.Payload == nil {
		webArticle.Payload = make(map[string]interface{}, 1)
	}
	webArticle.Payload[conf.PayloadKey] = zsc

	result := w.db.Save(webArticle)
	if result.Error != nil {
		return false, fmt.Errorf("saving Web Article: %w", result.Error)
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

func (w *Worker) connectGRPC() error {
	addr := w.config.ZeroShotClassification.ZeroShotGRPCAddress
	w.logger.Info().Msgf("Creating: gRPC connection [%s]", addr)
	conn, err := grpcutils.OpenConnection(addr, w.config.ZeroShotClassification.GRPCTLSDisable)
	if err != nil {
		w.logger.Fatal().Err(err).Msg("Failed to connect to gRPC API")
	}
	w.spagoGateway = NewGateway(conn)
	return nil
}
