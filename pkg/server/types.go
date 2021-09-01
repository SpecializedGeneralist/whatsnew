// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"database/sql"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server/whatsnew"
	"time"
)

func makeAPIFeed(feed models.Feed) *whatsnew.Feed {
	return &whatsnew.Feed{
		Id:              fmt.Sprintf("%d", feed.ID),
		CreatedAt:       feed.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       feed.UpdatedAt.Format(time.RFC3339),
		Url:             feed.URL,
		Enabled:         feed.Enabled,
		LastRetrievedAt: nullTimeToString(feed.LastRetrievedAt),
		FailuresCount:   int64(feed.FailuresCount),
		LastError:       feed.LastError.String,
	}
}

func makeAPIQueryTwitterSource(source models.TwitterSource) *whatsnew.QueryTwitterSource {
	return &whatsnew.QueryTwitterSource{
		Id:              fmt.Sprintf("%d", source.ID),
		CreatedAt:       source.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       source.UpdatedAt.Format(time.RFC3339),
		Query:           source.Text,
		Enabled:         source.Enabled,
		LastRetrievedAt: nullTimeToString(source.LastRetrievedAt),
		FailuresCount:   int64(source.FailuresCount),
		LastError:       source.LastError.String,
	}
}

func makeAPIUserTwitterSource(source models.TwitterSource) *whatsnew.UserTwitterSource {
	return &whatsnew.UserTwitterSource{
		Id:              fmt.Sprintf("%d", source.ID),
		CreatedAt:       source.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       source.UpdatedAt.Format(time.RFC3339),
		Username:        source.Text,
		Enabled:         source.Enabled,
		LastRetrievedAt: nullTimeToString(source.LastRetrievedAt),
		FailuresCount:   int64(source.FailuresCount),
		LastError:       source.LastError.String,
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
	return sql.NullTime{Time: t.UTC(), Valid: true}, nil
}
