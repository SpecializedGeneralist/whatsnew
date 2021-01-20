package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/zeroshotclassification"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) zeroShotClassification() *cli.Command {
	return &cli.Command{
		Name:  "zero-shot-classification",
		Usage: "Classify scraped news articles with a spaGO zero-shot classification service",
		Action: func(c *cli.Context) (err error) {
			db, err := models.OpenDB(app.config.DB.DSN, app.newContextLogger(c))
			if err != nil {
				return err
			}

			rmq := rabbitmq.NewClient(app.config.RabbitMQ.URI, app.config.RabbitMQ.ExchangeName)
			err = rmq.Connect()
			if err != nil {
				return err
			}
			defer func() {
				if e := rmq.Disconnect(); e != nil && err == nil {
					err = e
				}
			}()
			return zeroshotclassification.Classify(app.config, db, rmq, app.newContextLogger(c))
		},
	}
}
