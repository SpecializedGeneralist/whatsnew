// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package configuration

import (
	"github.com/rs/zerolog"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestFromYAMLFile(t *testing.T) {
	t.Parallel()

	t.Run("loading sample configuration file", func(t *testing.T) {
		t.Parallel()

		omitFeedItemsPublishedBefore, err := time.Parse(time.RFC3339, "2020-12-01T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}

		expected := Configuration{
			LogLevel: LogLevel(zerolog.InfoLevel),
			DB: DBConfiguration{
				DSN: "host=localhost port=5432 user=postgres password=postgres dbname=whatsnew sslmode=disable statement_cache_mode=describe",
			},
			RabbitMQ: RabbitMQConfiguration{
				URI:          "amqp://guest:guest@localhost:5672/",
				ExchangeName: "whatsnew",
			},
			FeedsFetching: FeedsFetchingConfiguration{
				NumWorkers:                          50,
				MaxAllowedFailures:                  15,
				SleepingTime:                        10 * time.Minute,
				OmitFeedItemsPublishedBeforeEnabled: true,
				OmitFeedItemsPublishedBefore:        omitFeedItemsPublishedBefore,
				NewWebResourceRoutingKey:            "new_web_resource",
				NewFeedItemRoutingKey:               "new_feed_item",
			},
			GDELTFetching: GDELTFetchingConfiguration{
				SleepingTime:                    5 * time.Minute,
				NewWebResourceRoutingKey:        "new_web_resource",
				NewGDELTEventRoutingKey:         "new_gdelt_event",
				TopLevelCameoEventCodeWhitelist: []string{},
			},
			WebScraping: WebScrapingConfiguration{
				NumWorkers:                  40,
				SubQueueName:                "whatsnew.web_scraping",
				SubNewWebResourceRoutingKey: "new_web_resource",
				PubNewWebArticleRoutingKey:  "new_web_article",
			},
			DuplicateDetector: DuplicateDetectorConfiguration{
				TimeframeHours:          72,
				SimilarityThreshold:     0.7,
				SubQueueName:            "whatsnew.duplicate_detector",
				SubRoutingKey:           "new_vectorized_web_article",
				PubNewEventRoutingKey:   "new_event",
				PubNewRelatedRoutingKey: "new_related",
			},
			SupportedLanguages: []string{"en", "es"},
		}

		filename := path.Join(getProjectRootDir(), "configuration", "testdata", "sample-configuration.yml")
		fileMustExist(t, filename)
		config, err := FromYAMLFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(config, expected) {
			t.Fatalf("expected %#v, actual %#v", expected, config)
		}
	})

	t.Run("loading a non-existent file", func(t *testing.T) {
		t.Parallel()

		filename := path.Join(getProjectRootDir(), "this-file-should-not-exist.foo")
		fileMustNotExist(t, filename)
		_, err := FromYAMLFile(filename)
		if err == nil {
			t.Fatal("expected error, actual nil")
		}
	})

	t.Run("loading a non-YAML file", func(t *testing.T) {
		t.Parallel()

		filename := path.Join(getProjectRootDir(), "configuration", "testdata", "another_file.txt")
		fileMustExist(t, filename)
		_, err := FromYAMLFile(filename)
		if err == nil {
			t.Fatal("expected error, actual nil")
		}
	})
}

func getProjectRootDir() string {
	_, testFilename, _, ok := runtime.Caller(0)
	if !ok {
		panic("error getting current test filename")
	}
	return path.Join(path.Dir(testFilename), "..")
}

func fileMustExist(t *testing.T, filename string) {
	t.Helper()
	if !fileExists(filename) {
		t.Fatalf("the file %#v must exist", filename)
	}
}

func fileMustNotExist(t *testing.T, filename string) {
	t.Helper()
	if fileExists(filename) {
		t.Fatalf("the file %#v must not exist", filename)
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
