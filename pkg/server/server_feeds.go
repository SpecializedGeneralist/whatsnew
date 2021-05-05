// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/api"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"gorm.io/gorm"
)

// GetFeeds gets all Feeds.
func (s *Server) GetFeeds(_ context.Context, req *api.GetFeedsRequest) (*api.GetFeedsResponse, error) {
	query := s.db.Unscoped().Order("id")
	if len(req.GetAfter()) > 0 {
		query = query.Where("id > ?", req.GetAfter())
	}
	if req.GetFirst() > 0 {
		query = query.Limit(int(req.GetFirst()))
	}

	var feeds []models.Feed
	ret := query.Find(&feeds)
	if ret.Error != nil {
		return &api.GetFeedsResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respFeeds := make([]*api.Feed, len(feeds))
	for i, feed := range feeds {
		respFeeds[i] = makeAPIFeed(feed)
	}

	resp := &api.GetFeedsResponse{
		Data: &api.GetFeedsData{
			Feeds: respFeeds,
		},
	}
	return resp, nil
}

// CreateFeeds creates new Feeds.
func (s *Server) CreateFeeds(_ context.Context, req *api.CreateFeedsRequest) (*api.CreateFeedsResponse, error) {
	reqFeeds := req.GetNewFeeds().GetFeeds()

	feeds := make([]models.Feed, len(reqFeeds))
	for i, reqFeed := range reqFeeds {
		feeds[i] = models.Feed{
			URL: reqFeed.GetUrl(),
		}
	}

	ret := s.db.Create(&feeds)
	if ret.Error != nil {
		return &api.CreateFeedsResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(feeds))
	for i, feed := range feeds {
		ids[i] = fmt.Sprintf("%d", feed.ID)
	}

	resp := &api.CreateFeedsResponse{
		Data: &api.CreateFeedsData{
			FeedIds: ids,
		},
	}
	return resp, nil
}

// CreateFeed creates a new Feed.
func (s *Server) CreateFeed(_ context.Context, req *api.CreateFeedRequest) (*api.CreateFeedResponse, error) {
	feed := models.Feed{
		URL: req.GetNewFeed().GetUrl(),
	}
	ret := s.db.Create(&feed)
	if ret.Error != nil {
		return &api.CreateFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &api.CreateFeedResponse{
		Data: &api.CreateFeedData{
			FeedId: fmt.Sprintf("%d", feed.ID),
		},
	}
	return resp, nil
}

// GetFeed gets a Feed.
func (s *Server) GetFeed(_ context.Context, req *api.GetFeedRequest) (*api.GetFeedResponse, error) {
	var feed models.Feed
	ret := s.db.Unscoped().First(&feed, "id = ?", req.GetId())
	if ret.Error != nil {
		return &api.GetFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &api.GetFeedResponse{
		Data: &api.GetFeedData{
			Feed: makeAPIFeed(feed),
		},
	}
	return resp, nil
}

// UpdateFeed updates a Feed.
func (s *Server) UpdateFeed(_ context.Context, req *api.UpdateFeedRequest) (*api.UpdateFeedResponse, error) {
	var feed models.Feed
	ret := s.db.Unscoped().First(&feed, "id = ?", req.GetId())
	if ret.Error != nil {
		return &api.UpdateFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	uf := req.GetUpdatedFeed()

	feed.URL = uf.GetUrl()
	feed.FailuresCount = int(uf.GetFailuresCount())

	deletedAt, err := nullTimeFromString(uf.GetDeletedAt())
	if err != nil {
		return &api.UpdateFeedResponse{Errors: s.makeErrors(req, err)}, nil
	}
	feed.DeletedAt = gorm.DeletedAt(deletedAt)

	feed.LastRetrievedAt, err = nullTimeFromString(uf.GetLastRetrievedAt())
	if err != nil {
		return &api.UpdateFeedResponse{Errors: s.makeErrors(req, err)}, nil
	}

	ret = s.db.Save(feed)
	if ret.Error != nil {
		return &api.UpdateFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	resp := &api.UpdateFeedResponse{
		Data: &api.UpdateFeedData{
			Feed: makeAPIFeed(feed),
		},
	}
	return resp, nil
}

// DeleteFeed deletes a Feed.
func (s *Server) DeleteFeed(_ context.Context, req *api.DeleteFeedRequest) (*api.DeleteFeedResponse, error) {
	var feed models.Feed
	ret := s.db.Unscoped().First(&feed, "id = ?", req.GetId())
	if ret.Error != nil {
		return &api.DeleteFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ret = s.db.Unscoped().Delete(&feed)
	if ret.Error != nil {
		return &api.DeleteFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &api.DeleteFeedResponse{
		Data: &api.DeleteFeedData{
			DeletedFeedId: fmt.Sprintf("%d", feed.ID),
		},
	}
	return resp, nil
}
