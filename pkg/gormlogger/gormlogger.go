// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gormlogger

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

// New returns a new logger for GORM messages.
func New(zl zerolog.Logger) logger.Interface {
	return &gormLogger{zl: zl}
}

var gormLevelToZerologLevel = map[logger.LogLevel]zerolog.Level{
	logger.Silent: zerolog.NoLevel,
	logger.Error:  zerolog.ErrorLevel,
	logger.Warn:   zerolog.WarnLevel,
	logger.Info:   zerolog.InfoLevel,
}

// LogMode returns a new logger configured with the given log level.
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &gormLogger{zl: l.zl.Level(gormLevelToZerologLevel[level])}
}

// Info prints an info message.
func (l gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	log.Ctx(l.zl.WithContext(ctx)).Info().Msgf(msg, data...)
}

// Warn prints an info message.
func (l gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log.Ctx(l.zl.WithContext(ctx)).Warn().Msgf(msg, data...)
}

// Error prints an info message.
func (l gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	log.Ctx(l.zl.WithContext(ctx)).Error().Msgf(msg, data...)
}

// Trace print sql message
func (l gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	switch {
	case err != nil && l.zl.GetLevel() <= zerolog.ErrorLevel:
		sql, rows := fc()
		log.Ctx(l.zl.WithContext(ctx)).Trace().Err(err).Str("file", utils.FileWithLineNum()).
			Str("sql", sql).Int64("rows", rows).Send()
	case l.zl.GetLevel() <= zerolog.TraceLevel:
		sql, rows := fc()
		log.Ctx(l.zl.WithContext(ctx)).Trace().Err(err).Str("file", utils.FileWithLineNum()).
			Str("sql", sql).Int64("rows", rows).Send()
	}
}
