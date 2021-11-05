// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import "github.com/SpecializedGeneralist/whatsnew/pkg/server/whatsnew"

func (s *Server) makeErrors(req interface{}, err error) *whatsnew.ResponseErrors {
	s.log.Debug().Err(err).Interface("request", req).Send()
	return &whatsnew.ResponseErrors{
		Value: []*whatsnew.ResponseError{
			{Message: err.Error()},
		},
	}
}
