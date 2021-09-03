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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetFeeds gets all Feeds.
func (s *Server) GetFeeds(_ context.Context, req *whatsnew.GetFeedsRequest) (*whatsnew.GetFeedsResponse, error) {
	query := s.db.Order("id")
	if len(req.GetAfter()) > 0 {
		query = query.Where("id > ?", req.GetAfter())
	}
	if req.GetFirst() > 0 {
		query = query.Limit(int(req.GetFirst()))
	}

	var feeds []models.Feed
	ret := query.Find(&feeds)
	if ret.Error != nil {
		return &whatsnew.GetFeedsResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respFeeds := make([]*whatsnew.Feed, len(feeds))
	for i, feed := range feeds {
		respFeeds[i] = makeAPIFeed(feed)
	}

	resp := &whatsnew.GetFeedsResponse{
		Data: &whatsnew.GetFeedsData{
			Feeds: respFeeds,
		},
	}
	return resp, nil
}

// CreateFeeds creates new Feeds.
func (s *Server) CreateFeeds(
	_ context.Context,
	req *whatsnew.CreateFeedsRequest,
) (*whatsnew.CreateFeedsResponse, error) {
	reqFeeds := req.GetNewFeeds().GetFeeds()

	feeds := make([]models.Feed, len(reqFeeds))
	for i, reqFeed := range reqFeeds {
		feeds[i] = models.Feed{
			URL: reqFeed.GetUrl(),
		}
	}

	ret := s.db.Create(&feeds)
	if ret.Error != nil {
		return &whatsnew.CreateFeedsResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(feeds))
	for i, feed := range feeds {
		ids[i] = fmt.Sprintf("%d", feed.ID)
	}

	resp := &whatsnew.CreateFeedsResponse{
		Data: &whatsnew.CreateFeedsData{
			FeedIds: ids,
		},
	}
	return resp, nil
}

// CreateFeed creates a new Feed.
func (s *Server) CreateFeed(_ context.Context, req *whatsnew.CreateFeedRequest) (*whatsnew.CreateFeedResponse, error) {
	feed := models.Feed{
		URL: req.GetNewFeed().GetUrl(),
	}
	ret := s.db.Create(&feed)
	if ret.Error != nil {
		return &whatsnew.CreateFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.CreateFeedResponse{
		Data: &whatsnew.CreateFeedData{
			FeedId: fmt.Sprintf("%d", feed.ID),
		},
	}
	return resp, nil
}

// GetFeed gets a Feed.
func (s *Server) GetFeed(_ context.Context, req *whatsnew.GetFeedRequest) (*whatsnew.GetFeedResponse, error) {
	var feed models.Feed
	ret := s.db.First(&feed, "id = ?", req.GetId())
	if ret.Error != nil {
		return &whatsnew.GetFeedResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.GetFeedResponse{
		Data: &whatsnew.GetFeedData{
			Feed: makeAPIFeed(feed),
		},
	}
	return resp, nil
}

// UpdateFeed updates a Feed.
func (s *Server) UpdateFeed(
	ctx context.Context,
	req *whatsnew.UpdateFeedRequest,
) (*whatsnew.UpdateFeedResponse, error) {
	var feed models.Feed

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&feed, "id = ?", req.GetId())
		if ret.Error != nil {
			return ret.Error
		}

		uf := req.GetUpdatedFeed()

		feed.URL = uf.GetUrl()
		feed.Enabled = uf.GetEnabled()
		feed.FailuresCount = int(uf.GetFailuresCount())
		feed.LastError = sql.NullString{
			String: uf.GetLastError(),
			Valid:  len(uf.GetLastError()) > 0,
		}

		var err error
		feed.LastRetrievedAt, err = nullTimeFromString(uf.GetLastRetrievedAt())
		if err != nil {
			return err
		}

		ret = tx.Save(&feed)
		return ret.Error
	})

	if err != nil {
		return &whatsnew.UpdateFeedResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.UpdateFeedResponse{
		Data: &whatsnew.UpdateFeedData{
			Feed: makeAPIFeed(feed),
		},
	}
	return resp, nil
}

// DeleteFeed deletes a Feed.
func (s *Server) DeleteFeed(
	ctx context.Context,
	req *whatsnew.DeleteFeedRequest,
) (*whatsnew.DeleteFeedResponse, error) {
	var feed models.Feed

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&feed, "id = ?", req.GetId())
		if ret.Error != nil {
			return ret.Error
		}

		var feedItemsCount int64
		ret = tx.Model(&models.FeedItem{}).Where("feed_id = ?", feed.ID).Limit(1).Count(&feedItemsCount)
		if ret.Error != nil {
			return ret.Error
		}

		if feedItemsCount == 0 {
			ret = tx.Unscoped().Delete(&feed)
		} else {
			ret = tx.Delete(&feed)
		}
		return ret.Error
	})

	if err != nil {
		return &whatsnew.DeleteFeedResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.DeleteFeedResponse{
		Data: &whatsnew.DeleteFeedData{
			DeletedFeedId: fmt.Sprintf("%d", feed.ID),
		},
	}
	return resp, nil
}
