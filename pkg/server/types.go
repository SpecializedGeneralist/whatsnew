// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/api"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"time"
)

func makeAPIFeed(feed models.Feed) *api.Feed {
	return &api.Feed{
		Id:              fmt.Sprintf("%d", feed.ID),
		CreatedAt:       feed.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       feed.UpdatedAt.Format(time.RFC3339),
		DeletedAt:       nullTimeToString(sql.NullTime(feed.DeletedAt)),
		Url:             feed.URL,
		LastRetrievedAt: nullTimeToString(feed.LastRetrievedAt),
		FailuresCount:   int64(feed.FailuresCount),
	}
}

func makeAPIQueryTwitterSource(source models.TwitterSource) *api.QueryTwitterSource {
	return &api.QueryTwitterSource{
		Id:              fmt.Sprintf("%d", source.ID),
		CreatedAt:       source.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       source.UpdatedAt.Format(time.RFC3339),
		Query:           source.Value,
		LastRetrievedAt: nullTimeToString(source.LastRetrievedAt),
	}
}

func makeAPIUserTwitterSource(source models.TwitterSource) *api.UserTwitterSource {
	return &api.UserTwitterSource{
		Id:              fmt.Sprintf("%d", source.ID),
		CreatedAt:       source.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       source.UpdatedAt.Format(time.RFC3339),
		Username:        source.Value,
		LastRetrievedAt: nullTimeToString(source.LastRetrievedAt),
	}
}

func nullTimeToString(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

func nullTimeFromString(s string) (sql.NullTime, error) {
	if len(s) == 0 {
		return sql.NullTime{Valid: false}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return sql.NullTime{Valid: false}, fmt.Errorf("error parsing RFC3339 datetime %#v: %w", s, err)
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}
