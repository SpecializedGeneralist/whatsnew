// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server/whatsnew"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Server implements WhatsNew HTTP and gRPC server.
type Server struct {
	whatsnew.UnimplementedWhatsnewServer
	conf config.Server
	db   *gorm.DB
	log  zerolog.Logger
}

// New creates a new Server.
func New(conf config.Server, db *gorm.DB) *Server {
	return &Server{
		conf: conf,
		db:   db,
		log:  log.Logger.Level(zerolog.Level(conf.LogLevel)),
	}
}
