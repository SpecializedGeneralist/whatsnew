// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package duplicatedetector

import (
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks"
	"github.com/jackc/pgtype"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"time"
)

type Worker struct {
	config         configuration.Configuration
	db             *gorm.DB
	rmq            *rabbitmq.Client
	customizeQuery CustomizeQueryFunc
	logger         zerolog.Logger
}

func (w *Worker) do(delivery amqp.Delivery) {
	w.logger.Debug().Msgf("processing msg %v", delivery.MessageId)

	webArticleID, err := rabbitmq.DecodeIDMessage(delivery.Body)
	if err != nil {
		w.logger.Err(err).Msg("error decoding ID message")
		w.sendNack(delivery)
		return
	}

	err = w.processWebArticleID(webArticleID)
	if err != nil {
		w.logger.Err(err).Msg("error processing web article ID")
		w.sendNack(delivery)
		return
	}

	w.sendAck(delivery)
}

func (w *Worker) processWebArticleID(webArticleID uint) error {
	webArticle, err := models.FindWebArticle(w.db, webArticleID)
	if err != nil {
		return fmt.Errorf("getting web article %d, %v", webArticleID, err)
	}

	if webArticle.Vector.Status != pgtype.Present {
		return fmt.Errorf("web article %d does not have a Vector", webArticleID)
	}

	return w.processWebArticle(webArticle)
}

func (w *Worker) processWebArticle(webArticle *models.WebArticle) error {
	maxID, maxScore, err := w.findMostSimilarWebArticle(webArticle)

	if err != nil {
		return fmt.Errorf("find most similar acticle to %d error: %v", webArticle.ID, err)
	}

	logger := w.logger.With().Uint("WebArticleID", webArticle.ID).
		Uint("maxID", maxID).Float32("maxScore", maxScore).Logger()

	if maxID == 0 || maxScore < w.config.DuplicateDetector.SimilarityThreshold {
		err = w.rmq.PublishID(w.config.DuplicateDetector.PubNewEventRoutingKey, webArticle.ID)
		if err != nil {
			return fmt.Errorf("error publishing new event %d: %v", webArticle.ID, err)
		}

		logger.Info().Msg("new event found")
		return nil
	}

	logger.Info().Msg("near-duplicate article found")

	webArticle.RelatedToWebArticleID = &maxID
	webArticle.RelatedScore = sql.NullFloat64{Float64: float64(maxScore), Valid: true}
	result := w.db.Save(webArticle)
	if result.Error != nil {
		return fmt.Errorf("error saving related article %d: %v", webArticle.ID, err)
	}

	err = w.rmq.PublishID(w.config.DuplicateDetector.PubNewRelatedRoutingKey, webArticle.ID)
	if err != nil {
		return fmt.Errorf("error publishing new related %d: %v", webArticle.ID, err)
	}

	return nil
}

func (w *Worker) findMostSimilarWebArticle(webArticle *models.WebArticle) (uint, float32, error) {
	referenceVector, err := tasks.ByteSliceToFloat32Slice(webArticle.Vector.Bytes)
	if err != nil {
		return 0, 0, fmt.Errorf("reference vector byteSliceToFloat32Slice: %v", err)
	}

	timeframe := time.Duration(w.config.DuplicateDetector.TimeframeHours) * time.Hour
	pastDate := webArticle.PublishDate.Add(-timeframe)

	query := w.db.
		Model(&models.WebArticle{}).
		Where("vector IS NOT NULL").
		Where("id < ?", webArticle.ID).
		Where("publish_date >= ?", pastDate.Format(time.RFC3339))

	if w.customizeQuery != nil {
		query, err = w.customizeQuery(query, webArticle)
		if err != nil {
			return 0, 0, fmt.Errorf("query customization error: %v", err)
		}
	}

	var maxScore float32
	var maxID uint = 0

	var results []models.WebArticle
	result := query.Select("id", "vector").FindInBatches(&results, 100, func(tx *gorm.DB, batch int) error {
		if tx.Error != nil {
			return tx.Error
		}
		for _, result := range results {
			otherVector, err := tasks.ByteSliceToFloat32Slice(result.Vector.Bytes)
			if err != nil {
				return fmt.Errorf("other vector byteSliceToFloat32Slice: %v", err)
			}
			score := cosineSimilarity(referenceVector, otherVector)

			if maxID == 0 || maxScore < score {
				maxID = result.ID
				maxScore = score
			}
		}
		return nil
	})

	if result.Error != nil {
		return 0, 0, fmt.Errorf("find in batches error: %v", query.Error)
	}

	return maxID, maxScore, nil
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

func cosineSimilarity(v1, v2 []float32) float32 {
	var s float32 = 0
	_ = v2[len(v1)-1]
	for i, a := range v1 {
		s += a * v2[i]
	}
	return s
}
