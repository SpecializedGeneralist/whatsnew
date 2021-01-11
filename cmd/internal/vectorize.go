package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/vectorizer"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) vectorize() *cli.Command {
	return &cli.Command{
		Name:  "vectorize",
		Usage: "Set `WebArticle.Vector` through LaBSE encoding",
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

			return vectorizer.Vectorize(app.config, db, rmq, app.newContextLogger(c))
		},
	}
}
