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

// GetQueryTwitterSources gets all Query Twitter Sources.
func (s *Server) GetQueryTwitterSources(_ context.Context, req *api.GetQueryTwitterSourcesRequest) (*api.GetQueryTwitterSourcesResponse, error) {
	query := s.db.Order("id").Where("type = ?", models.SearchTwitterSource)
	if len(req.GetAfter()) > 0 {
		query = query.Where("id > ?", req.GetAfter())
	}
	if req.GetFirst() > 0 {
		query = query.Limit(int(req.GetFirst()))
	}

	var sources []models.TwitterSource
	ret := query.Find(&sources)
	if ret.Error != nil {
		return &api.GetQueryTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respSources := make([]*api.QueryTwitterSource, len(sources))
	for i, source := range sources {
		respSources[i] = makeAPIQueryTwitterSource(source)
	}

	resp := &api.GetQueryTwitterSourcesResponse{
		Data: &api.GetQueryTwitterSourcesData{
			QueryTwitterSources: respSources,
		},
	}
	return resp, nil
}

// CreateQueryTwitterSources creates new Query Twitter Sources.
func (s *Server) CreateQueryTwitterSources(_ context.Context, req *api.CreateQueryTwitterSourcesRequest) (*api.CreateQueryTwitterSourcesResponse, error) {
	reqSources := req.GetNewQueryTwitterSources().GetQueryTwitterSources()
	sources := make([]models.TwitterSource, len(reqSources))
	for i, reqSource := range reqSources {
		sources[i] = models.TwitterSource{
			Type:  models.SearchTwitterSource,
			Value: reqSource.GetQuery(),
		}
	}

	ret := s.db.Create(&sources)
	if ret.Error != nil {
		return &api.CreateQueryTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(sources))
	for i, source := range sources {
		ids[i] = fmt.Sprintf("%d", source.ID)
	}

	resp := &api.CreateQueryTwitterSourcesResponse{
		Data: &api.CreateQueryTwitterSourcesData{
			QueryTwitterSourceIds: ids,
		},
	}
	return resp, nil
}

// CreateQueryTwitterSource creates a new Query Twitter Source.
func (s *Server) CreateQueryTwitterSource(_ context.Context, req *api.CreateQueryTwitterSourceRequest) (*api.CreateQueryTwitterSourceResponse, error) {
	ts := &models.TwitterSource{
		Type:  models.SearchTwitterSource,
		Value: req.GetNewQueryTwitterSource().GetQuery(),
	}

	ret := s.db.Create(ts)
	if ret.Error != nil {
		return &api.CreateQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &api.CreateQueryTwitterSourceResponse{
		Data: &api.CreateQueryTwitterSourceData{
			QueryTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}

// GetQueryTwitterSource gets all Query Twitter Sources.
func (s *Server) GetQueryTwitterSource(_ context.Context, req *api.GetQueryTwitterSourceRequest) (*api.GetQueryTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.SearchTwitterSource, req.GetId())
	if ret.Error != nil {
		return &api.GetQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &api.GetQueryTwitterSourceResponse{
		Data: &api.GetQueryTwitterSourceData{
			QueryTwitterSource: makeAPIQueryTwitterSource(ts),
		},
	}
	return resp, nil
}

// UpdateQueryTwitterSource updates a Query Twitter Source.
func (s *Server) UpdateQueryTwitterSource(_ context.Context, req *api.UpdateQueryTwitterSourceRequest) (*api.UpdateQueryTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.SearchTwitterSource, req.GetId())
	if ret.Error != nil {
		return &api.UpdateQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	us := req.GetUpdatedQueryTwitterSource()

	ts.Value = us.GetQuery()

	var err error
	ts.LastRetrievedAt, err = nullTimeFromString(us.GetLastRetrievedAt())
	if err != nil {
		return &api.UpdateQueryTwitterSourceResponse{Errors: s.makeErrors(req, err)}, nil
	}

	ret = s.db.Save(ts)
	if ret.Error != nil {
		return &api.UpdateQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &api.UpdateQueryTwitterSourceResponse{
		Data: &api.UpdateQueryTwitterSourceData{
			QueryTwitterSource: makeAPIQueryTwitterSource(ts),
		},
	}
	return resp, nil
}

// DeleteQueryTwitterSource deletes a Query Twitter Source.
func (s *Server) DeleteQueryTwitterSource(_ context.Context, req *api.DeleteQueryTwitterSourceRequest) (*api.DeleteQueryTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.SearchTwitterSource, req.GetId())
	if ret.Error != nil {
		return &api.DeleteQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ret = s.db.Delete(&ts)
	if ret.Error != nil {
		return &api.DeleteQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &api.DeleteQueryTwitterSourceResponse{
		Data: &api.DeleteQueryTwitterSourceData{
			DeletedQueryTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}
