// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	gormlogger "gorm.io/gorm/logger"
	"os"
	"strings"
	"time"
)

// Config holds whatsnew application-wide configuration settings.
type Config struct {
	DB              DB              `yaml:"db"`
	Faktory         Faktory         `yaml:"faktory"`
	FeedsScheduling FeedsScheduling `yaml:"feeds_scheduling"`
	Workers         Workers         `yaml:"workers"`
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
	Queues   []string `yaml:"queues"`
	LogLevel LogLevel `yaml:"loglevel"`
}

// FeedsScheduling holds settings for scheduling feeds for further processing.
type FeedsScheduling struct {
	TimeInterval time.Duration `yaml:"time_interval"`
	Jobs         []string      `yaml:"jobs"`
	LogLevel     LogLevel      `yaml:"loglevel"`
}

// Workers holds settings for the various workers.
type Workers struct {
	FeedFetcher FeedFetcher `yaml:"feed_fetcher"`
}

// FeedFetcher holds settings for the FeedFetcher worker.
type FeedFetcher struct {
	Concurrency              int                      `yaml:"concurrency"`
	NewWebResourceJobs       []string                 `yaml:"new_web_resource_jobs"`
	MaxAllowedFailures       int                      `yaml:"max_allowed_failures"`
	OmitItemsPublishedBefore OmitItemsPublishedBefore `yaml:"omit_items_published_before"`
	LanguageFilter           []string                 `yaml:"language_filter"`
	LogLevel                 LogLevel                 `yaml:"loglevel"`
}

// OmitItemsPublishedBefore is part of FeedFetcher settings.
type OmitItemsPublishedBefore struct {
	Enabled bool      `yaml:"enabled"`
	Time    time.Time `yaml:"time"`
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
