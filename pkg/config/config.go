// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	hnswgrpcapi "github.com/SpecializedGeneralist/hnsw-grpc-server/pkg/grpcapi"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	gormlogger "gorm.io/gorm/logger"
	"os"
	"regexp"
	"strings"
	"time"
)

// Config holds whatsnew application-wide configuration settings.
type Config struct {
	DB               DB               `yaml:"db"`
	Faktory          Faktory          `yaml:"faktory"`
	HNSW             HNSW             `yaml:"hnsw"`
	FeedScheduler    FeedScheduler    `yaml:"feed_scheduler"`
	TwitterScheduler TwitterScheduler `yaml:"twitter_scheduler"`
	GDELTFetcher     GDELTFetcher     `yaml:"gdelt_fetcher"`
	JobsRecoverer    JobsRecoverer    `yaml:"jobs_recoverer"`
	Server           Server           `yaml:"server"`
	Workers          Workers          `yaml:"workers"`
}

// DB holds database settings.
type DB struct {
	// DSN, dbname excluded.
	DSN      string     `yaml:"dsn"`
	DBName   string     `yaml:"dbname"`
	LogLevel DBLogLevel `yaml:"loglevel"`
}

// Faktory holds Faktory settings and generic workers properties.
type Faktory struct {
	URL      string   `yaml:"url"`
	LogLevel LogLevel `yaml:"loglevel"`
}

// HNSW holds settings for connecting to HNSW server and handling vector indices.
type HNSW struct {
	Server GRPCServer `yaml:"server"`
	Index  HNSWIndex  `yaml:"index"`
}

// HNSWIndex holds settings for HNSW vector indices.
type HNSWIndex struct {
	NamePrefix     string        `yaml:"name_prefix"`
	Dim            int32         `yaml:"dim"`
	EfConstruction int32         `yaml:"ef_construction"`
	M              int32         `yaml:"m"`
	MaxElements    int32         `yaml:"max_elements"`
	Seed           int32         `yaml:"seed"`
	SpaceType      HNSWSpaceType `yaml:"space_type"`
}

// FeedScheduler holds settings for scheduling feeds for further processing.
type FeedScheduler struct {
	TimeInterval time.Duration `yaml:"time_interval"`
	Jobs         []FaktoryJob  `yaml:"jobs"`
	LogLevel     LogLevel      `yaml:"loglevel"`
}

// TwitterScheduler holds settings for scheduling twitter sources for further
// processing.
type TwitterScheduler struct {
	TimeInterval time.Duration `yaml:"time_interval"`
	Jobs         []FaktoryJob  `yaml:"jobs"`
	LogLevel     LogLevel      `yaml:"loglevel"`
}

// GDELTFetcher holds settings for fetching GDELT events and extracting news
// report URLs for further processing.
type GDELTFetcher struct {
	TimeInterval           time.Duration `yaml:"time_interval"`
	EventRootCodeWhitelist []string      `yaml:"event_root_code_whitelist"`
	NewWebResourceJobs     []FaktoryJob  `yaml:"new_web_resource_jobs"`
	LogLevel               LogLevel      `yaml:"loglevel"`
}

// JobsRecoverer holds settings for the periodic recovery process of
// pending jobs.
type JobsRecoverer struct {
	TimeInterval time.Duration `yaml:"time_interval"`
	LeewayTime   time.Duration `yaml:"leeway_time"`
	LogLevel     LogLevel      `yaml:"loglevel"`
}

// Server holds settings for the HTTP and gRPC server.
type Server struct {
	Address        string   `yaml:"address"`
	TLSEnabled     bool     `yaml:"tls_enabled"`
	TLSCert        string   `yaml:"tls_cert"`
	TLSKey         string   `yaml:"tls_key"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	LogLevel       LogLevel `yaml:"loglevel"`
}

// Workers holds settings for the various workers.
type Workers struct {
	FeedFetcher          FeedFetcher          `yaml:"feed_fetcher"`
	TwitterScraper       TwitterScraper       `yaml:"twitter_scraper"`
	WebScraper           WebScraper           `yaml:"web_scraper"`
	Translator           Translator           `yaml:"translator"`
	ZeroShotClassifier   ZeroShotClassifier   `yaml:"zero_shot_classifier"`
	Vectorizer           Vectorizer           `yaml:"vectorizer"`
	DuplicateDetector    DuplicateDetector    `yaml:"duplicate_detector"`
	InformationExtractor InformationExtractor `yaml:"information_extractor"`
}

// FeedFetcher holds settings for the FeedFetcher worker.
type FeedFetcher struct {
	Queues                   []string                 `yaml:"queues"`
	Concurrency              int                      `yaml:"concurrency"`
	NewWebResourceJobs       []FaktoryJob             `yaml:"new_web_resource_jobs"`
	MaxAllowedFailures       int                      `yaml:"max_allowed_failures"`
	OmitItemsPublishedBefore OmitItemsPublishedBefore `yaml:"omit_items_published_before"`
	LanguageFilter           []string                 `yaml:"language_filter"`
	LogLevel                 LogLevel                 `yaml:"loglevel"`
}

// TwitterScraper holds settings for the TwitterScraper worker.
type TwitterScraper struct {
	Queues                    []string                 `yaml:"queues"`
	Concurrency               int                      `yaml:"concurrency"`
	MaxTweetsNumber           int                      `yaml:"max_tweets_number"`
	NewWebArticleJobs         []FaktoryJob             `yaml:"new_web_article_jobs"`
	OmitTweetsPublishedBefore OmitItemsPublishedBefore `yaml:"omit_tweets_published_before"`
	LanguageFilter            []string                 `yaml:"language_filter"`
	LogLevel                  LogLevel                 `yaml:"loglevel"`
}

// WebScraper holds settings for the WebScraper worker.
type WebScraper struct {
	Queues            []string      `yaml:"queues"`
	Concurrency       int           `yaml:"concurrency"`
	NewWebArticleJobs []FaktoryJob  `yaml:"new_web_article_jobs"`
	LanguageFilter    []string      `yaml:"language_filter"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
	UserAgent         string        `yaml:"user_agent"`
	LogLevel          LogLevel      `yaml:"loglevel"`
}

// Translator holds settings for the translator worker.
type Translator struct {
	Queues                  []string     `yaml:"queues"`
	Concurrency             int          `yaml:"concurrency"`
	TranslatorServer        GRPCServer   `yaml:"translator_server"`
	ProcessedWebArticleJobs []FaktoryJob `yaml:"processed_web_article_jobs"`
	LanguageWhitelist       []string     `yaml:"language_whitelist"`
	TargetLanguage          string       `yaml:"target_language"`
	LogLevel                LogLevel     `yaml:"loglevel"`
}

// ZeroShotClassifier holds settings for the zero-shot classifier worker.
type ZeroShotClassifier struct {
	Queues                  []string     `yaml:"queues"`
	Concurrency             int          `yaml:"concurrency"`
	ProcessedWebArticleJobs []FaktoryJob `yaml:"processed_web_article_jobs"`
	SpagoBARTServer         GRPCServer   `yaml:"spago_bart_server"`
	LogLevel                LogLevel     `yaml:"loglevel"`
}

// Vectorizer holds settings for the Vectorizer worker.
type Vectorizer struct {
	Queues                   []string     `yaml:"queues"`
	Concurrency              int          `yaml:"concurrency"`
	VectorizedWebArticleJobs []FaktoryJob `yaml:"vectorized_web_article_jobs"`
	SpagoBERTServer          GRPCServer   `yaml:"spago_bert_server"`
	LogLevel                 LogLevel     `yaml:"loglevel"`
}

// DuplicateDetector holds settings for the duplicate detector worker.
type DuplicateDetector struct {
	Queues                     []string     `yaml:"queues"`
	TimeframeDays              int          `yaml:"timeframe_days"`
	DistanceThreshold          float32      `yaml:"distance_threshold"`
	NonDuplicateWebArticleJobs []FaktoryJob `yaml:"non_duplicate_web_article_jobs"`
	DuplicateWebArticleJobs    []FaktoryJob `yaml:"duplicate_web_article_jobs"`
	LogLevel                   LogLevel     `yaml:"loglevel"`
}

// InformationExtractor holds settings for the information extractor worker.
type InformationExtractor struct {
	Queues                  []string     `yaml:"queues"`
	Concurrency             int          `yaml:"concurrency"`
	SpagoBERTServer         GRPCServer   `yaml:"spago_bert_server"`
	ProcessedWebArticleJobs []FaktoryJob `yaml:"processed_web_article_jobs"`
	LogLevel                LogLevel     `yaml:"loglevel"`
}

// OmitItemsPublishedBefore is part of FeedFetcher settings.
type OmitItemsPublishedBefore struct {
	Enabled bool      `yaml:"enabled"`
	Time    time.Time `yaml:"time"`
}

// GRPCServer holds common settings for connecting to a gRPC server.
type GRPCServer struct {
	Target     string `yaml:"target"`
	TLSEnabled bool   `yaml:"tls_enabled"`
}

// FaktoryJob describes a Faktory job to be scheduled for execution.
type FaktoryJob struct {
	JobType string `yaml:"job_type"`
	Queue   string `yaml:"queue"`
}

// DBLogLevel is a redefinition of GORM logger.LogLevel which satisfies
// encoding.TextUnmarshaler, to be conveniently parsed from YAML.
type DBLogLevel gormlogger.LogLevel

var dbLogLevels = map[string]DBLogLevel{
	"silent": DBLogLevel(gormlogger.Silent),
	"error":  DBLogLevel(gormlogger.Error),
	"warn":   DBLogLevel(gormlogger.Warn),
	"info":   DBLogLevel(gormlogger.Info),
}

// UnmarshalText satisfies the encoding.TextUnmarshaler interface, unmarshaling
// the text to a DBLogLevel.
func (l *DBLogLevel) UnmarshalText(text []byte) error {
	s := string(text)
	level, ok := dbLogLevels[s]
	if !ok {
		return fmt.Errorf("invalid DB log level: %#v", s)
	}
	*l = level
	return nil
}

// LogLevel is a redefinition of zerolog.Level which satisfies
// encoding.TextUnmarshaler, to be conveniently parsed from YAML.
type LogLevel zerolog.Level

// UnmarshalText satisfies the encoding.TextUnmarshaler interface, unmarshaling
// the text to a LogLevel.
func (l *LogLevel) UnmarshalText(text []byte) (err error) {
	s := string(text)
	zl, err := zerolog.ParseLevel(s)
	if err != nil || zl == zerolog.NoLevel {
		return fmt.Errorf("invalid log level: %#v", s)
	}
	*l = LogLevel(zl)
	return nil
}

// HNSWSpaceType is a redefinition of HNSW gRPC API CreateIndexRequest_SpaceType
// which satisfies encoding.TextUnmarshaler, to be conveniently parsed from YAML.
type HNSWSpaceType hnswgrpcapi.CreateIndexRequest_SpaceType

// UnmarshalText satisfies the encoding.TextUnmarshaler interface, unmarshaling
// the text to an HNSW gRPC API CreateIndexRequest_SpaceType.
func (hst *HNSWSpaceType) UnmarshalText(text []byte) (err error) {
	s := string(text)
	st, ok := hnswgrpcapi.CreateIndexRequest_SpaceType_value[s]
	if !ok {
		return fmt.Errorf("invalid HNSW space type: %#v", s)
	}
	*hst = HNSWSpaceType(st)
	return nil
}

// Regexp is a redefinition of regexp.Regexp which satisfies
// encoding.TextUnmarshaler, to be conveniently parsed from YAML.
type Regexp regexp.Regexp

// UnmarshalText satisfies the encoding.TextUnmarshaler interface, unmarshaling
// the text to a Regexp.
func (r *Regexp) UnmarshalText(text []byte) (err error) {
	s := string(text)
	re, err := regexp.Compile(s)
	if err != nil {
		return fmt.Errorf("invalid regular expression %#v: %w", s, err)
	}

	*r = Regexp(*re)
	return nil
}

// FromYAMLFile reads a Config object from a YAML file.
//
// Before being decoded, the whole YAML file content is passed through
// os.ExpandEnv.
func FromYAMLFile(filename string) (*Config, error) {
	rawContent, err := os.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("cannot read config file %#v: %w", filename, err)
		return nil, err
	}
	content := os.ExpandEnv(string(rawContent))

	conf := new(Config)
	err = yaml.NewDecoder(strings.NewReader(content)).Decode(conf)
	if err != nil {
		err = fmt.Errorf("cannot decode config file %#v: %w", filename, err)
		return nil, err
	}
	return conf, nil
}
