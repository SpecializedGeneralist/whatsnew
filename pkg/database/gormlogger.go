// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type gormLogger struct {
	zl zerolog.Logger
}

// NewGORMLogger returns a new logger for GORM.
//
// The logger implementation uses zerolog under the hood, and it is initialized
// with the global zerolog Logger. A zerolog-compatible log level is also set
// from the global logger.
func NewGORMLogger() logger.Interface {
	return &gormLogger{zl: log.Logger}
}

var gormLevelToZerologLevel = map[logger.LogLevel]zerolog.Level{
	logger.Silent: zerolog.Disabled,
	logger.Error:  zerolog.ErrorLevel,
	logger.Warn:   zerolog.WarnLevel,
	logger.Info:   zerolog.InfoLevel,
}

// LogMode returns a new logger configured with the given log level.
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &gormLogger{zl: l.zl.Level(gormLevelToZerologLevel[level])}
}

// Info prints an info message.
func (l gormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	l.zl.Info().Msgf(msg, data...)
}

// Warn prints an info message.
func (l gormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	l.zl.Warn().Msgf(msg, data...)
}

// Error prints an info message.
func (l gormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	l.zl.Error().Msgf(msg, data...)
}

// Trace print sql message
func (l gormLogger) Trace(_ context.Context, _ time.Time, fc func() (string, int64), err error) {
	level := l.zl.GetLevel()
	switch {
	case err != nil && level <= zerolog.ErrorLevel:
		sql, rows := fc()
		l.zl.Error().Err(err).Str("file", utils.FileWithLineNum()).
			Str("sql", sql).Int64("rows", rows).Send()
	case level <= zerolog.InfoLevel:
		sql, rows := fc()
		l.zl.Info().Err(err).Str("file", utils.FileWithLineNum()).
			Str("sql", sql).Int64("rows", rows).Send()
	}
}
