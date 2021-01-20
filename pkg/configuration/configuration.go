// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

// Configuration provides app-wide settings.
type Configuration struct {
	LogLevel               LogLevel `yaml:"log_level"`
	DB                     DBConfiguration
	RabbitMQ               RabbitMQConfiguration
	FeedsFetching          FeedsFetchingConfiguration          `yaml:"feeds_fetching"`
	GDELTFetching          GDELTFetchingConfiguration          `yaml:"gdelt_fetching"`
	WebScraping            WebScrapingConfiguration            `yaml:"web_scraping"`
	DuplicateDetector      DuplicateDetectorConfiguration      `yaml:"duplicate_detector"`
	Vectorizer             VectorizerConfiguration             `yaml:"vectorizer"`
	ZeroShotClassification ZeroShotClassificationConfiguration `yaml:"zero_shot_classification"`
	SupportedLanguages     []string                            `yaml:"supported_languages"`
}

// DBConfiguration provides database-specific settings.
type DBConfiguration struct {
	DSN string
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

// DuplicateDetectorConfiguration provides specific settings for near-duplicate web
// articles detection.
type DuplicateDetectorConfiguration struct {
	TimeframeHours          int     `yaml:"timeframe_hours"`
	SimilarityThreshold     float32 `yaml:"similarity_threshold"`
	SubQueueName            string  `yaml:"sub_queue_name"`
	SubRoutingKey           string  `yaml:"sub_routing_key"`
	PubNewEventRoutingKey   string  `yaml:"pub_new_event_routing_key"`
	PubNewRelatedRoutingKey string  `yaml:"pub_new_related_routing_key"`
}

// VectorizerConfiguration provides specific settings for the vectorization operation.
type VectorizerConfiguration struct {
	NumWorkers                           int    `yaml:"num_workers"`
	SubQueueName                         string `yaml:"sub_queue_name"`
	SubNewWebArticleRoutingKey           string `yaml:"sub_new_web_article_routing_key"`
	PubNewVectorizedWebArticleRoutingKey string `yaml:"pub_new_vectorized_web_article_routing_key"`
	LabseGrpcAddress                     string `yaml:"labse_grpc_address"`
	LabseTLSDisable                      bool   `yaml:"labse_tls_disable"`
}

// ZeroShotClassificationConfiguration provides specific settings for spaGO zero-shot classification operation.
type ZeroShotClassificationConfiguration struct {
	NumWorkers          int      `yaml:"num_workers"`
	SubQueueName        string   `yaml:"sub_queue_name"`
	SubRoutingKey       string   `yaml:"sub_routing_key"`
	PubRoutingKey       string   `yaml:"pub_routing_key"`
	PayloadKey          string   `yaml:"payload_key"`
	ZeroShotGRPCAddress string   `yaml:"zero_shot_grpc_address"`
	GRPCTLSDisable      bool     `yaml:"grpc_tls_disable"`
	HypothesisTemplate  string   `yaml:"hypothesis_template"`
	PossibleLabels      []string `yaml:"possible_labels"`
	MultiClass          bool     `yaml:"multi_class"`
}

func (c *Configuration) LanguageIsSupported(code string) bool {
	for _, c := range c.SupportedLanguages {
		if code == c {
			return true
		}
	}
	return false
}

// LogLevel is a redefinition of zerolog.Level which satisfies encoding.TextUnmarshaler.
type LogLevel zerolog.Level

// UnmarshalText unmarshals the text to a LogLevel.
func (l *LogLevel) UnmarshalText(text []byte) (err error) {
	zl, err := zerolog.ParseLevel(string(text))
	*l = LogLevel(zl)
	return err
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
