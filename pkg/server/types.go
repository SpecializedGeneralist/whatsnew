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

func makeAPIZeroShotHypothesisTemplate(t models.ZeroShotHypothesisTemplate) *whatsnew.ZeroShotHypothesisTemplate {
	return &whatsnew.ZeroShotHypothesisTemplate{
		Id:         fmt.Sprintf("%d", t.ID),
		CreatedAt:  t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  t.UpdatedAt.Format(time.RFC3339),
		Enabled:    t.Enabled,
		Text:       t.Text,
		MultiClass: t.MultiClass,
		Labels:     makeAPIZeroShotHypothesisLabels(t.Labels),
	}
}

func makeAPIZeroShotHypothesisLabels(labels []models.ZeroShotHypothesisLabel) []*whatsnew.ZeroShotHypothesisLabel {
	out := make([]*whatsnew.ZeroShotHypothesisLabel, len(labels))
	for i, label := range labels {
		out[i] = makeAPIZeroShotHypothesisLabel(label)
	}
	return out
}

func makeAPIZeroShotHypothesisLabel(label models.ZeroShotHypothesisLabel) *whatsnew.ZeroShotHypothesisLabel {
	return &whatsnew.ZeroShotHypothesisLabel{
		Id:        fmt.Sprintf("%d", label.ID),
		CreatedAt: label.CreatedAt.Format(time.RFC3339),
		UpdatedAt: label.UpdatedAt.Format(time.RFC3339),
		Enabled:   label.Enabled,
		Text:      label.Text,
	}
}

func makeAPIInfoExtractionRule(rule models.InfoExtractionRule) *whatsnew.InfoExtractionRule {
	return &whatsnew.InfoExtractionRule{
		Id:           fmt.Sprintf("%d", rule.ID),
		CreatedAt:    rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    rule.UpdatedAt.Format(time.RFC3339),
		Label:        rule.Label,
		Question:     rule.Question,
		AnswerRegexp: rule.AnswerRegexp.String(),
		Threshold:    rule.Threshold,
		Enabled:      rule.Enabled,
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
