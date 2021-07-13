// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config_test

import (
	"fmt"
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
				Queues:   []string{"default"},
				LogLevel: config.LogLevel(zerolog.InfoLevel),
			},
			FeedScheduler: config.FeedScheduler{
				TimeInterval: 5 * time.Minute,
				Jobs:         []string{"FeedFetcher"},
				LogLevel:     config.LogLevel(zerolog.InfoLevel),
			},
			TwitterScheduler: config.TwitterScheduler{
				TimeInterval: 5 * time.Minute,
				Jobs:         []string{"TwitterScraper"},
				LogLevel:     config.LogLevel(zerolog.InfoLevel),
			},
			GDELTFetcher: config.GDELTFetcher{
				TimeInterval:           5 * time.Minute,
				EventRootCodeWhitelist: make([]string, 0),
				NewWebResourceJobs:     []string{"WebScraper"},
				LogLevel:               config.LogLevel(zerolog.InfoLevel),
			},
			Workers: config.Workers{
				FeedFetcher: config.FeedFetcher{
					Concurrency:        10,
					NewWebResourceJobs: []string{"WebScraper"},
					MaxAllowedFailures: 15,
					OmitItemsPublishedBefore: config.OmitItemsPublishedBefore{
						Enabled: true,
						Time:    time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC),
					},
					LanguageFilter: []string{"en", "es"},
					LogLevel:       config.LogLevel(zerolog.InfoLevel),
				},
				TwitterScraper: config.TwitterScraper{
					Concurrency:       10,
					MaxTweetsNumber:   1000,
					NewWebArticleJobs: []string{"Vectorizer"},
					OmitTweetsPublishedBefore: config.OmitItemsPublishedBefore{
						Enabled: true,
						Time:    time.Date(2021, time.July, 1, 0, 0, 0, 0, time.UTC),
					},
					LanguageFilter: []string{"en", "es"},
					LogLevel:       config.LogLevel(zerolog.InfoLevel),
				},
				WebScraper: config.WebScraper{
					Concurrency:       10,
					NewWebArticleJobs: []string{"Vectorizer"},
					LanguageFilter:    []string{"en", "es"},
					RequestTimeout:    30 * time.Second,
					UserAgent:         "WhatsNew/0.0.0",
					LogLevel:          config.LogLevel(zerolog.InfoLevel),
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

func dataFile(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func sampleConfigFile() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "sample-config.yml")
}
