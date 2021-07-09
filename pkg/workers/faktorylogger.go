// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package workers

import (
	"fmt"
	"github.com/contribsys/faktory_worker_go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type faktoryLogger struct {
	zl zerolog.Logger
}

// NewFaktoryLogger returns a new logger for Faktory Manager.
//
// The logger implementation uses zerolog under the hood.
// It is initialized with a logger derived from the global zerolog Logger,
// just setting the loglevel to the given value.
func NewFaktoryLogger(level zerolog.Level) faktory_worker.Logger {
	return &faktoryLogger{zl: log.Logger.Level(level)}
}

func (fl *faktoryLogger) Debug(v ...interface{}) {
	fl.zl.Debug().Msgf(fmt.Sprint(v...))
}

func (fl *faktoryLogger) Debugf(format string, args ...interface{}) {
	fl.zl.Debug().Msgf(format, args...)
}

func (fl *faktoryLogger) Info(v ...interface{}) {
	fl.zl.Info().Msgf(fmt.Sprint(v...))
}

func (fl *faktoryLogger) Infof(format string, args ...interface{}) {
	fl.zl.Info().Msgf(format, args...)
}

func (fl *faktoryLogger) Warn(v ...interface{}) {
	fl.zl.Warn().Msgf(fmt.Sprint(v...))
}

func (fl *faktoryLogger) Warnf(format string, args ...interface{}) {
	fl.zl.Warn().Msgf(format, args...)
}

func (fl *faktoryLogger) Error(v ...interface{}) {
	fl.zl.Error().Msgf(fmt.Sprint(v...))
}

func (fl *faktoryLogger) Errorf(format string, args ...interface{}) {
	fl.zl.Error().Msgf(format, args...)
}

func (fl *faktoryLogger) Fatal(v ...interface{}) {
	fl.zl.Fatal().Msgf(fmt.Sprint(v...))
}

func (fl *faktoryLogger) Fatalf(format string, args ...interface{}) {
	fl.zl.Fatal().Msgf(format, args...)
}
