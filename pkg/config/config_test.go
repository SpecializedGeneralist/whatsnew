// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config_test

import (
	"fmt"
	hnswgrpcapi "github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlogger "gorm.io/gorm/logger"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestFromYAMLFile(t *testing.T) {
	t.Parallel()

	t.Run("empty filename", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile("")
		assert.Error(t, err)
		assert.Nil(t, conf)
	})

	t.Run("missing filename", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(dataFile("foo"))
		assert.Error(t, err)
		assert.Nil(t, conf)
	})

	t.Run("malformed YAML", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(dataFile("malformed-config.yml"))
		assert.Error(t, err)
		assert.Nil(t, conf)
	})

	t.Run("empty YAML", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(dataFile("empty-config.yml"))
		assert.NoError(t, err)
		require.NotNil(t, conf)
		assert.Equal(t, config.Config{}, *conf)
	})

	t.Run("sample-config.yml", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(sampleConfigFile())
		assert.NoError(t, err)
		require.NotNil(t, conf)
		expected := config.Config{
			DB: config.DB{
				DSN:      "host=localhost port=5432 user=postgres password=postgres sslmode=disable statement_cache_mode=describe",
				DBName:   "whatsnew",
				LogLevel: config.DBLogLevel(gormlogger.Warn),
			},
			Faktory: config.Faktory{
				URL:      "tcp://faktory:faktory@localhost:7419",
				LogLevel: config.LogLevel(zerolog.InfoLevel),
			},
			HNSW: config.HNSW{
				Server: config.GRPCServer{
					Target:     "127.0.0.1:19530",
					TLSEnabled: false,
				},
				Index: config.HNSWIndex{
					NamePrefix:     "whatsnew_",
					Dim:            768,
					EfConstruction: 200,
					M:              48,
					MaxElements:    100000,
					Seed:           42,
					SpaceType:      config.HNSWSpaceType(hnswgrpcapi.CreateIndexRequest_COSINE),
				},
			},
			Server: config.Server{
				Address:        "0.0.0.0:10000",
				TLSEnabled:     false,
				TLSCert:        "",
				TLSKey:         "",
				AllowedOrigins: []string{"*"},
				LogLevel:       config.LogLevel(zerolog.InfoLevel),
			},
			Tasks: config.Tasks{
				FeedScheduler: config.FeedScheduler{
					TimeInterval: 5 * time.Minute,
					Jobs: []config.FaktoryJob{
						{
							JobType:    "FeedFetcher",
							Queue:      "feed_fetcher",
							ReserveFor: 300,
							Retry:      -1,
						},
					},
					LogLevel: config.LogLevel(zerolog.InfoLevel),
				},
				TwitterScheduler: config.TwitterScheduler{
					TimeInterval: 5 * time.Minute,
					Jobs: []config.FaktoryJob{
						{
							JobType:    "TwitterScraper",
							Queue:      "twitter_scraper",
							ReserveFor: 300,
							Retry:      -1,
						},
					},
					LogLevel: config.LogLevel(zerolog.InfoLevel),
				},
				GDELTFetcher: config.GDELTFetcher{
					TimeInterval:           5 * time.Minute,
					EventRootCodeWhitelist: make([]string, 0),
					NewWebResourceJobs: []config.FaktoryJob{
						{
							JobType:    "WebScraper",
							Queue:      "web_scraper",
							ReserveFor: 600,
							Retry:      5,
						},
					},
					LogLevel: config.LogLevel(zerolog.InfoLevel),
				},
				JobsRecoverer: config.JobsRecoverer{
					TimeInterval: time.Minute,
					LeewayTime:   time.Minute,
					LogLevel:     config.LogLevel(zerolog.InfoLevel),
				},
				HNSWPurger: config.HNSWPurger{
					TimeInterval:               1 * time.Hour,
					DeleteIndicesOlderThanDays: 6,
					LogLevel:                   config.LogLevel(zerolog.InfoLevel),
				},
			},
			Workers: config.Workers{
				FeedFetcher: config.FeedFetcher{
					Queues:      []string{"feed_fetcher"},
					Concurrency: 10,
					NewWebResourceJobs: []config.FaktoryJob{
						{
							JobType:    "WebScraper",
							Queue:      "web_scraper",
							ReserveFor: 600,
							Retry:      5,
						},
					},
					MaxAllowedFailures: 15,
					OmitItemsPublishedBefore: config.OmitItemsPublishedBefore{
						Enabled: true,
						Time:    time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC),
					},
					LanguageFilter: []string{"en", "es", "fr", "it"},
					RequestTimeout: 10 * time.Second,
					LogLevel:       config.LogLevel(zerolog.InfoLevel),
				},
				TwitterScraper: config.TwitterScraper{
					Queues:          []string{"twitter_scraper"},
					Concurrency:     10,
					MaxTweetsNumber: 1000,
					NewWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "Translator",
							Queue:      "translator",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					OmitTweetsPublishedBefore: config.OmitItemsPublishedBefore{
						Enabled: true,
						Time:    time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC),
					},
					LanguageFilter: []string{"en", "es", "fr", "it"},
					LogLevel:       config.LogLevel(zerolog.InfoLevel),
				},
				WebScraper: config.WebScraper{
					Queues:      []string{"web_scraper"},
					Concurrency: 10,
					NewWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "Translator",
							Queue:      "translator",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					LanguageFilter: []string{"en", "es", "fr", "it"},
					RequestTimeout: 30 * time.Second,
					UserAgent:      "WhatsNew/1.0.0-beta.3",
					LogLevel:       config.LogLevel(zerolog.InfoLevel),
				},
				Translator: config.Translator{
					Queues:      []string{"translator"},
					Concurrency: 4,
					TranslatorServer: config.GRPCServer{
						Target:     "127.0.0.1:4557",
						TLSEnabled: false,
					},
					ProcessedWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "ZeroShotClassifier",
							Queue:      "zero_shot_classifier",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					LanguageWhitelist: []string{"fr", "it"},
					TargetLanguage:    "en",
					LogLevel:          config.LogLevel(zerolog.InfoLevel),
				},
				ZeroShotClassifier: config.ZeroShotClassifier{
					Queues:      []string{"zero_shot_classifier"},
					Concurrency: 4,
					ProcessedWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "TextClassifier",
							Queue:      "text_classifier",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					SpagoBARTServer: config.GRPCServer{
						Target:     "127.0.0.1:4001",
						TLSEnabled: false,
					},
					LogLevel: config.LogLevel(zerolog.InfoLevel),
				},
				TextClassifier: config.TextClassifier{
					Queues:      []string{"text_classifier"},
					Concurrency: 4,
					ProcessedWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "GeoParser",
							Queue:      "geo_parser",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					ClassifierServer: config.GRPCServer{
						Target:     "127.0.0.1:4002",
						TLSEnabled: false,
					},
					LogLevel: config.LogLevel(zerolog.InfoLevel),
				},
				GeoParser: config.GeoParser{
					Queues:      []string{"geo_parser"},
					Concurrency: 4,
					ProcessedWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "Vectorizer",
							Queue:      "vectorizer",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					CliffURI: "http://127.0.0.1:4003",
					LogLevel: config.LogLevel(zerolog.InfoLevel),
				},
				Vectorizer: config.Vectorizer{
					Queues:      []string{"vectorizer"},
					Concurrency: 4,
					VectorizedWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "DuplicateDetector",
							Queue:      "duplicate_detector",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					SpagoBERTServer: config.GRPCServer{
						Target:     "127.0.0.1:1976",
						TLSEnabled: false,
					},
					LogLevel: config.LogLevel(zerolog.InfoLevel),
				},
				DuplicateDetector: config.DuplicateDetector{
					Queues:            []string{"duplicate_detector"},
					TimeframeDays:     3,
					DistanceThreshold: 0.3,
					NonDuplicateWebArticleJobs: []config.FaktoryJob{
						{
							JobType:    "InformationExtractor",
							Queue:      "information_extractor",
							ReserveFor: 600,
							Retry:      25,
						},
					},
					DuplicateWebArticleJobs: []config.FaktoryJob{},
					LogLevel:                config.LogLevel(zerolog.InfoLevel),
				},
				InformationExtractor: config.InformationExtractor{
					Queues:      []string{"information_extractor"},
					Concurrency: 4,
					SpagoBERTServer: config.GRPCServer{
						Target:     "127.0.0.1:5831",
						TLSEnabled: false,
					},
					ProcessedWebArticleJobs: []config.FaktoryJob{},
					LogLevel:                config.LogLevel(zerolog.InfoLevel),
				},
			},
		}
		assert.Equal(t, expected, *conf)
	})
}

func TestDBLogLevel_UnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("positive cases", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			text     string
			expected config.DBLogLevel
		}{
			{"silent", config.DBLogLevel(gormlogger.Silent)},
			{"error", config.DBLogLevel(gormlogger.Error)},
			{"warn", config.DBLogLevel(gormlogger.Warn)},
			{"info", config.DBLogLevel(gormlogger.Info)},
		}
		for _, tc := range testCases {
			t.Run(tc.text, func(t *testing.T) {
				l := new(config.DBLogLevel)
				err := l.UnmarshalText([]byte(tc.text))
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, *l)
			})
		}
	})

	t.Run("negative cases", func(t *testing.T) {
		t.Parallel()
		testCases := []string{
			"",
			" ",
			"foo",
		}
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%#v", tc), func(t *testing.T) {
				l := new(config.DBLogLevel)
				err := l.UnmarshalText([]byte(tc))
				assert.Error(t, err)
			})
		}
	})
}

func TestLogLevel_UnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("positive cases", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			text     string
			expected config.LogLevel
		}{
			{"trace", config.LogLevel(zerolog.TraceLevel)},
			{"debug", config.LogLevel(zerolog.DebugLevel)},
			{"info", config.LogLevel(zerolog.InfoLevel)},
			{"warn", config.LogLevel(zerolog.WarnLevel)},
			{"error", config.LogLevel(zerolog.ErrorLevel)},
			{"fatal", config.LogLevel(zerolog.FatalLevel)},
			{"panic", config.LogLevel(zerolog.PanicLevel)},
			{"disabled", config.LogLevel(zerolog.Disabled)},
		}
		for _, tc := range testCases {
			t.Run(tc.text, func(t *testing.T) {
				l := new(config.LogLevel)
				err := l.UnmarshalText([]byte(tc.text))
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, *l)
			})
		}
	})

	t.Run("negative cases", func(t *testing.T) {
		t.Parallel()
		testCases := []string{
			"",
			" ",
			"foo",
		}
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%#v", tc), func(t *testing.T) {
				l := new(config.LogLevel)
				err := l.UnmarshalText([]byte(tc))
				assert.Error(t, err)
			})
		}
	})
}

func TestHNSWSpaceType_UnmarshalText(t *testing.T) {
	t.Parallel()

	t.Run("positive cases", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			text     string
			expected config.HNSWSpaceType
		}{
			{"L2", config.HNSWSpaceType(hnswgrpcapi.CreateIndexRequest_L2)},
			{"IP", config.HNSWSpaceType(hnswgrpcapi.CreateIndexRequest_IP)},
			{"COSINE", config.HNSWSpaceType(hnswgrpcapi.CreateIndexRequest_COSINE)},
		}
		for _, tc := range testCases {
			t.Run(tc.text, func(t *testing.T) {
				l := new(config.HNSWSpaceType)
				err := l.UnmarshalText([]byte(tc.text))
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, *l)
			})
		}
	})

	t.Run("negative cases", func(t *testing.T) {
		t.Parallel()
		testCases := []string{
			"",
			" ",
			"foo",
		}
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%#v", tc), func(t *testing.T) {
				l := new(config.HNSWSpaceType)
				err := l.UnmarshalText([]byte(tc))
				assert.Error(t, err)
			})
		}
	})
}

func dataFile(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func sampleConfigFile() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "sample-config.yml")
}
