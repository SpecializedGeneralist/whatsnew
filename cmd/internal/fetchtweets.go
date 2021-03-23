// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/rabbitmq"
	"github.com/SpecializedGeneralist/whatsnew/pkg/tasks/tweetsfetching"
	"github.com/urfave/cli/v2"
)

func (app *CLIApp) fetchTweets() *cli.Command {
	return &cli.Command{
		Name:  "fetch-tweets",
		Usage: "Fetch all tweets",
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

			return tweetsfetching.FetchTweets(app.config, db, rmq, app.newContextLogger(c))
		},
	}
}
