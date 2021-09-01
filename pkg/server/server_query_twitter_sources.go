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

// GetQueryTwitterSources gets all Query Twitter Sources.
func (s *Server) GetQueryTwitterSources(_ context.Context, req *whatsnew.GetQueryTwitterSourcesRequest) (*whatsnew.GetQueryTwitterSourcesResponse, error) {
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
		return &whatsnew.GetQueryTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respSources := make([]*whatsnew.QueryTwitterSource, len(sources))
	for i, source := range sources {
		respSources[i] = makeAPIQueryTwitterSource(source)
	}

	resp := &whatsnew.GetQueryTwitterSourcesResponse{
		Data: &whatsnew.GetQueryTwitterSourcesData{
			QueryTwitterSources: respSources,
		},
	}
	return resp, nil
}

// CreateQueryTwitterSources creates new Query Twitter Sources.
func (s *Server) CreateQueryTwitterSources(_ context.Context, req *whatsnew.CreateQueryTwitterSourcesRequest) (*whatsnew.CreateQueryTwitterSourcesResponse, error) {
	reqSources := req.GetNewQueryTwitterSources().GetQueryTwitterSources()
	sources := make([]models.TwitterSource, len(reqSources))
	for i, reqSource := range reqSources {
		sources[i] = models.TwitterSource{
			Type: models.SearchTwitterSource,
			Text: reqSource.GetQuery(),
		}
	}

	ret := s.db.Create(&sources)
	if ret.Error != nil {
		return &whatsnew.CreateQueryTwitterSourcesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(sources))
	for i, source := range sources {
		ids[i] = fmt.Sprintf("%d", source.ID)
	}

	resp := &whatsnew.CreateQueryTwitterSourcesResponse{
		Data: &whatsnew.CreateQueryTwitterSourcesData{
			QueryTwitterSourceIds: ids,
		},
	}
	return resp, nil
}

// CreateQueryTwitterSource creates a new Query Twitter Source.
func (s *Server) CreateQueryTwitterSource(_ context.Context, req *whatsnew.CreateQueryTwitterSourceRequest) (*whatsnew.CreateQueryTwitterSourceResponse, error) {
	ts := &models.TwitterSource{
		Type: models.SearchTwitterSource,
		Text: req.GetNewQueryTwitterSource().GetQuery(),
	}

	ret := s.db.Create(ts)
	if ret.Error != nil {
		return &whatsnew.CreateQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &whatsnew.CreateQueryTwitterSourceResponse{
		Data: &whatsnew.CreateQueryTwitterSourceData{
			QueryTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}

// GetQueryTwitterSource gets all Query Twitter Sources.
func (s *Server) GetQueryTwitterSource(_ context.Context, req *whatsnew.GetQueryTwitterSourceRequest) (*whatsnew.GetQueryTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.SearchTwitterSource, req.GetId())
	if ret.Error != nil {
		return &whatsnew.GetQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.GetQueryTwitterSourceResponse{
		Data: &whatsnew.GetQueryTwitterSourceData{
			QueryTwitterSource: makeAPIQueryTwitterSource(ts),
		},
	}
	return resp, nil
}

// UpdateQueryTwitterSource updates a Query Twitter Source.
func (s *Server) UpdateQueryTwitterSource(_ context.Context, req *whatsnew.UpdateQueryTwitterSourceRequest) (*whatsnew.UpdateQueryTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.SearchTwitterSource, req.GetId())
	if ret.Error != nil {
		return &whatsnew.UpdateQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	us := req.GetUpdatedQueryTwitterSource()

	ts.Text = us.GetQuery()
	ts.Enabled = us.GetEnabled()
	ts.FailuresCount = int(us.GetFailuresCount())
	ts.LastError = sql.NullString{
		String: us.GetLastError(),
		Valid:  len(us.GetLastError()) > 0,
	}

	var err error
	ts.LastRetrievedAt, err = nullTimeFromString(us.GetLastRetrievedAt())
	if err != nil {
		return &whatsnew.UpdateQueryTwitterSourceResponse{Errors: s.makeErrors(req, err)}, nil
	}

	ret = s.db.Save(ts)
	if ret.Error != nil {
		return &whatsnew.UpdateQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &whatsnew.UpdateQueryTwitterSourceResponse{
		Data: &whatsnew.UpdateQueryTwitterSourceData{
			QueryTwitterSource: makeAPIQueryTwitterSource(ts),
		},
	}
	return resp, nil
}

// DeleteQueryTwitterSource deletes a Query Twitter Source.
func (s *Server) DeleteQueryTwitterSource(_ context.Context, req *whatsnew.DeleteQueryTwitterSourceRequest) (*whatsnew.DeleteQueryTwitterSourceResponse, error) {
	var ts models.TwitterSource
	ret := s.db.First(&ts, "type = ? AND id = ?", models.SearchTwitterSource, req.GetId())
	if ret.Error != nil {
		return &whatsnew.DeleteQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	var tweetsCount int64
	ret = s.db.Model(&models.Tweet{}).Where("twitter_source_id = ?", ts.ID).Limit(1).Count(&tweetsCount)
	if ret.Error != nil {
		return &whatsnew.DeleteQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	if tweetsCount == 0 {
		ret = s.db.Unscoped().Delete(&ts)
	} else {
		ret = s.db.Delete(&ts)
	}
	if ret.Error != nil {
		return &whatsnew.DeleteQueryTwitterSourceResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.DeleteQueryTwitterSourceResponse{
		Data: &whatsnew.DeleteQueryTwitterSourceData{
			DeletedQueryTwitterSourceId: fmt.Sprintf("%d", ts.ID),
		},
	}
	return resp, nil
}
