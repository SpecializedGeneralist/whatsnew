// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"encoding"
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type LogLevel struct {
	zerolog.Level
}

var _ encoding.TextUnmarshaler = &LogLevel{}

func (l *LogLevel) UnmarshalText(text []byte) (err error) {
	l.Level, err = zerolog.ParseLevel(string(text))
	return err
}

// Configuration provides app-wide settings.
type Configuration struct {
	LogLevel           LogLevel `yaml:"log_level"`
	DB                 DBConfiguration
	Elasticsearch      ElasticsearchConfiguration
	RabbitMQ           RabbitMQConfiguration
	FeedsFetching      FeedsFetchingConfiguration `yaml:"feeds_fetching"`
	GDELTFetching      GDELTFetchingConfiguration `yaml:"gdelt_fetching"`
	WebScraping        WebScrapingConfiguration   `yaml:"web_scraping"`
	SupportedLanguages []string                   `yaml:"supported_languages"`
}

// DBConfiguration provides database-specific settings.
type DBConfiguration struct {
	DSN string
}

// ElasticsearchConfiguration provides Elasticsearch specific settings (legacy DB).
type ElasticsearchConfiguration struct {
	URL       string `yaml:"url"`
	IndexName string `yaml:"index_name"`
}

// RabbitMQConfiguration provides RabbitMQ-specific settings.
type RabbitMQConfiguration struct {
	URI          string
	ExchangeName string `yaml:"exchange_name"`
}

// FeedsFetchingConfiguration provides specific settings for the
// feeds-fetching operation.
type FeedsFetchingConfiguration struct {
	NumWorkers                          int           `yaml:"num_workers"`
	MaxAllowedFailures                  int           `yaml:"max_allowed_failures"`
	SleepingTime                        time.Duration `yaml:"sleeping_time"`
	OmitFeedItemsPublishedBeforeEnabled bool          `yaml:"omit_feed_items_published_before_enabled"`
	OmitFeedItemsPublishedBefore        time.Time     `yaml:"omit_feed_items_published_before"`
	NewWebResourceRoutingKey            string        `yaml:"new_web_resource_routing_key"`
	NewFeedItemRoutingKey               string        `yaml:"new_feed_item_routing_key"`
}

// GDELTFetchingConfiguration provides specific settings for the
// GDELT-fetching operation.
type GDELTFetchingConfiguration struct {
	SleepingTime                    time.Duration `yaml:"sleeping_time"`
	NewWebResourceRoutingKey        string        `yaml:"new_web_resource_routing_key"`
	NewGDELTEventRoutingKey         string        `yaml:"new_gdelt_event_routing_key"`
	TopLevelCameoEventCodeWhitelist []string      `yaml:"top_level_cameo_event_code_whitelist"`
}

// WebScrapingConfiguration provides specific settings for the
// Web Resource URLs scraping operation.
type WebScrapingConfiguration struct {
	NumWorkers                  int    `yaml:"num_workers"`
	SubQueueName                string `yaml:"sub_queue_name"`
	SubNewWebResourceRoutingKey string `yaml:"sub_new_web_resource_routing_key"`
	PubNewWebArticleRoutingKey  string `yaml:"pub_new_web_article_routing_key"`
}

func (c *Configuration) LanguageIsSupported(code string) bool {
	for _, c := range c.SupportedLanguages {
		if code == c {
			return true
		}
	}
	return false
}

// FromYAMLFile reads a Configuration object from a YAML file.
func FromYAMLFile(filename string) (config Configuration, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return config, fmt.Errorf("open file %s: %v", filename, err)
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = fmt.Errorf("close file %s: %v", filename, e)
		}
	}()

	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return config, fmt.Errorf("decode YAML file %s: %v", filename, err)
	}
	return config, nil
}
