// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/api"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
)

// GetUserTwitterSources gets all User Twitter Sources.
func (s *Server) GetUserTwitterSources(_ context.Context, req *api.GetUserTwitterSourcesRequest) (*api.GetUserTwitterSourcesResponse, error) {
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
		return &api.GetUserTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respSources := make([]*api.UserTwitterSource, len(sources))
	for i, source := range sources {
		respSources[i] = makeAPIUserTwitterSource(source)
	}

	resp := &api.GetUserTwitterSourcesResponse{
		Data: &api.GetUserTwitterSourcesData{
			UserTwitterSources: respSources,
		},
	}
	return resp, nil
}

// CreateUserTwitterSources creates new User Twitter Sources.
func (s *Server) CreateUserTwitterSources(_ context.Context, req *api.CreateUserTwitterSourcesRequest) (*api.CreateUserTwitterSourcesResponse, error) {
	reqSources := req.GetNewUserTwitterSources().GetUserTwitterSources()
	sources := make([]models.TwitterSource, len(reqSources))
	for i, reqSource := range reqSources {
		sources[i] = models.TwitterSource{
			Type:  models.UserTwitterSource,
			Value: reqSource.GetUsername(),
		}
	}

	ret := s.db.Create(&sources)
	if ret.Error != nil {
		return &api.CreateUserTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(sources))
	for i, source := range sources {
		ids[i] = fmt.Sprintf("%d", source.ID)
	}

	resp := &api.CreateUserTwitterSourcesResponse{
		Data: &api.CreateUserTwitterSourcesData{
			UserTwitterSourceIds: ids,
		},
	}
	return resp, nil
}

// CreateUserTwitterSource creates a new User Twitter Source.
func (s *Server) CreateUserTwitterSource(_ context.Context, req *api.CreateUserTwitterSourceRequest) (*api.CreateUserTwitterSourceResponse, error) {
	ts := &models.TwitterSource{
		Type:  models.UserTwitterSource,
		Value: req.GetNewUserTwitterSource().GetUsername(),
	}

	ret := s.db.Create(ts)
	if ret.Error != nil {
		return &api.CreateUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &api.CreateUserTwitterSourceResponse{
		Data: &api.CreateUserTwitterSourceData{
			UserTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}

// GetUserTwitterSource gets a User Twitter Source.
func (s *Server) GetUserTwitterSource(_ context.Context, req *api.GetUserTwitterSourceRequest) (*api.GetUserTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.UserTwitterSource, req.GetId())
	if ret.Error != nil {
		return &api.GetUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &api.GetUserTwitterSourceResponse{
		Data: &api.GetUserTwitterSourceData{
			UserTwitterSource: makeAPIUserTwitterSource(ts),
		},
	}
	return resp, nil
}

// UpdateUserTwitterSource updates a User Twitter Source.
func (s *Server) UpdateUserTwitterSource(_ context.Context, req *api.UpdateUserTwitterSourceRequest) (*api.UpdateUserTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.UserTwitterSource, req.GetId())
	if ret.Error != nil {
		return &api.UpdateUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	us := req.GetUpdatedUserTwitterSource()

	ts.Value = us.GetUsername()

	var err error
	ts.LastRetrievedAt, err = nullTimeFromString(us.GetLastRetrievedAt())
	if err != nil {
		return &api.UpdateUserTwitterSourceResponse{Errors: s.makeErrors(req, err)}, nil
	}

	ret = s.db.Save(ts)
	if ret.Error != nil {
		return &api.UpdateUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &api.UpdateUserTwitterSourceResponse{
		Data: &api.UpdateUserTwitterSourceData{
			UserTwitterSource: makeAPIUserTwitterSource(ts),
		},
	}
	return resp, nil
}

// DeleteUserTwitterSource deletes a User Twitter Source.
func (s *Server) DeleteUserTwitterSource(_ context.Context, req *api.DeleteUserTwitterSourceRequest) (*api.DeleteUserTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.UserTwitterSource, req.GetId())
	if ret.Error != nil {
		return &api.DeleteUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ret = s.db.Delete(&ts)
	if ret.Error != nil {
		return &api.DeleteUserTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &api.DeleteUserTwitterSourceResponse{
		Data: &api.DeleteUserTwitterSourceData{
			DeletedUserTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}
