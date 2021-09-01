// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server/whatsnew"
)

// GetUserTwitterSources gets all User Twitter Sources.
func (s *Server) GetUserTwitterSources(_ context.Context, req *whatsnew.GetUserTwitterSourcesRequest) (*whatsnew.GetUserTwitterSourcesResponse, error) {
	query := s.db.Order("id").Where("type = ?", models.UserTwitterSource)
	if len(req.GetAfter()) > 0 {
		query = query.Where("id > ?", req.GetAfter())
	}
	if req.GetFirst() > 0 {
		query = query.Limit(int(req.GetFirst()))
	}

	var sources []models.TwitterSource
	ret := query.Find(&sources)
	if ret.Error != nil {
		return &whatsnew.GetUserTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respSources := make([]*whatsnew.UserTwitterSource, len(sources))
	for i, source := range sources {
		respSources[i] = makeAPIUserTwitterSource(source)
	}

	resp := &whatsnew.GetUserTwitterSourcesResponse{
		Data: &whatsnew.GetUserTwitterSourcesData{
			UserTwitterSources: respSources,
		},
	}
	return resp, nil
}

// CreateUserTwitterSources creates new User Twitter Sources.
func (s *Server) CreateUserTwitterSources(_ context.Context, req *whatsnew.CreateUserTwitterSourcesRequest) (*whatsnew.CreateUserTwitterSourcesResponse, error) {
	reqSources := req.GetNewUserTwitterSources().GetUserTwitterSources()
	sources := make([]models.TwitterSource, len(reqSources))
	for i, reqSource := range reqSources {
		sources[i] = models.TwitterSource{
			Type: models.UserTwitterSource,
			Text: reqSource.GetUsername(),
		}
	}

	ret := s.db.Create(&sources)
	if ret.Error != nil {
		return &whatsnew.CreateUserTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(sources))
	for i, source := range sources {
		ids[i] = fmt.Sprintf("%d", source.ID)
	}

	resp := &whatsnew.CreateUserTwitterSourcesResponse{
		Data: &whatsnew.CreateUserTwitterSourcesData{
			UserTwitterSourceIds: ids,
		},
	}
	return resp, nil
}

// CreateUserTwitterSource creates a new User Twitter Source.
func (s *Server) CreateUserTwitterSource(_ context.Context, req *whatsnew.CreateUserTwitterSourceRequest) (*whatsnew.CreateUserTwitterSourceResponse, error) {
	ts := &models.TwitterSource{
		Type: models.UserTwitterSource,
		Text: req.GetNewUserTwitterSource().GetUsername(),
	}

	ret := s.db.Create(ts)
	if ret.Error != nil {
		return &whatsnew.CreateUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &whatsnew.CreateUserTwitterSourceResponse{
		Data: &whatsnew.CreateUserTwitterSourceData{
			UserTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}

// GetUserTwitterSource gets a User Twitter Source.
func (s *Server) GetUserTwitterSource(_ context.Context, req *whatsnew.GetUserTwitterSourceRequest) (*whatsnew.GetUserTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.UserTwitterSource, req.GetId())
	if ret.Error != nil {
		return &whatsnew.GetUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.GetUserTwitterSourceResponse{
		Data: &whatsnew.GetUserTwitterSourceData{
			UserTwitterSource: makeAPIUserTwitterSource(ts),
		},
	}
	return resp, nil
}

// UpdateUserTwitterSource updates a User Twitter Source.
func (s *Server) UpdateUserTwitterSource(_ context.Context, req *whatsnew.UpdateUserTwitterSourceRequest) (*whatsnew.UpdateUserTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.UserTwitterSource, req.GetId())
	if ret.Error != nil {
		return &whatsnew.UpdateUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	us := req.GetUpdatedUserTwitterSource()

	ts.Text = us.GetUsername()
	ts.Enabled = us.GetEnabled()
	ts.FailuresCount = int(us.GetFailuresCount())
	ts.LastError = sql.NullString{
		String: us.GetLastError(),
		Valid:  len(us.GetLastError()) > 0,
	}

	var err error
	ts.LastRetrievedAt, err = nullTimeFromString(us.GetLastRetrievedAt())
	if err != nil {
		return &whatsnew.UpdateUserTwitterSourceResponse{Errors: s.makeErrors(req, err)}, nil
	}

	ret = s.db.Save(ts)
	if ret.Error != nil {
		return &whatsnew.UpdateUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &whatsnew.UpdateUserTwitterSourceResponse{
		Data: &whatsnew.UpdateUserTwitterSourceData{
			UserTwitterSource: makeAPIUserTwitterSource(ts),
		},
	}
	return resp, nil
}

// DeleteUserTwitterSource deletes a User Twitter Source.
func (s *Server) DeleteUserTwitterSource(_ context.Context, req *whatsnew.DeleteUserTwitterSourceRequest) (*whatsnew.DeleteUserTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.UserTwitterSource, req.GetId())
	if ret.Error != nil {
		return &whatsnew.DeleteUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ret = s.db.Delete(&ts)
	if ret.Error != nil {
		return &whatsnew.DeleteUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.DeleteUserTwitterSourceResponse{
		Data: &whatsnew.DeleteUserTwitterSourceData{
			DeletedUserTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}
