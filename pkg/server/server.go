// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/api"
	"github.com/SpecializedGeneralist/whatsnew/pkg/configuration"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Server is the main API server implementation.
type Server struct {
	api.UnimplementedApiServer
	config configuration.ServerConfiguration
	db     *gorm.DB
	logger zerolog.Logger
}

// New creates a new Server.
func New(config configuration.ServerConfiguration, db *gorm.DB, logger zerolog.Logger) *Server {
	return &Server{
		config: config,
		db:     db,
		logger: logger,
	}
}
