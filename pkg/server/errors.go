// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import "github.com/SpecializedGeneralist/whatsnew/pkg/api"

func (s *Server) makeErrors(req interface{}, err error) *api.ResponseErrors {
	s.logger.Trace().Err(err).Interface("request", req).Send()
	return &api.ResponseErrors{
		Value: []*api.ResponseError{
			{Message: err.Error()},
		},
	}
}

func (s *Server) makeFatalErrors(req interface{}, err error) *api.ResponseErrors {
	s.logger.Error().Err(err).Interface("request", req).Send()
	return &api.ResponseErrors{
		Value: []*api.ResponseError{
			{Message: err.Error()},
		},
	}
}
