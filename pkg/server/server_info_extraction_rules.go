// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models/types"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server/whatsnew"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"regexp"
)

// GetInfoExtractionRules gets all InfoExtractionRules.
func (s *Server) GetInfoExtractionRules(
	_ context.Context,
	req *whatsnew.GetInfoExtractionRulesRequest,
) (*whatsnew.GetInfoExtractionRulesResponse, error) {
	query := s.db.Order("id")
	if len(req.GetAfter()) > 0 {
		query = query.Where("id > ?", req.GetAfter())
	}
	if req.GetFirst() > 0 {
		query = query.Limit(int(req.GetFirst()))
	}

	var rules []models.InfoExtractionRule
	ret := query.Find(&rules)
	if ret.Error != nil {
		return &whatsnew.GetInfoExtractionRulesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respInfoExtractionRules := make([]*whatsnew.InfoExtractionRule, len(rules))
	for i, infoExtractionRule := range rules {
		respInfoExtractionRules[i] = makeAPIInfoExtractionRule(infoExtractionRule)
	}

	resp := &whatsnew.GetInfoExtractionRulesResponse{
		Data: &whatsnew.GetInfoExtractionRulesData{
			InfoExtractionRules: respInfoExtractionRules,
		},
	}
	return resp, nil
}

// CreateInfoExtractionRules creates new InfoExtractionRules.
func (s *Server) CreateInfoExtractionRules(
	_ context.Context,
	req *whatsnew.CreateInfoExtractionRulesRequest,
) (*whatsnew.CreateInfoExtractionRulesResponse, error) {
	reqRules := req.GetNewInfoExtractionRules().GetInfoExtractionRules()

	rules := make([]models.InfoExtractionRule, len(reqRules))
	for i, reqRule := range reqRules {
		model, err := makeInfoExtractionRuleModel(reqRule)
		if err != nil {
			return &whatsnew.CreateInfoExtractionRulesResponse{Errors: s.makeErrors(req, err)}, nil
		}
		rules[i] = *model
	}

	ret := s.db.Create(&rules)
	if ret.Error != nil {
		return &whatsnew.CreateInfoExtractionRulesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(rules))
	for i, rule := range rules {
		ids[i] = fmt.Sprintf("%d", rule.ID)
	}

	resp := &whatsnew.CreateInfoExtractionRulesResponse{
		Data: &whatsnew.CreateInfoExtractionRulesData{
			InfoExtractionRuleIds: ids,
		},
	}
	return resp, nil
}

// CreateInfoExtractionRule creates a new InfoExtractionRule.
func (s *Server) CreateInfoExtractionRule(
	_ context.Context,
	req *whatsnew.CreateInfoExtractionRuleRequest,
) (*whatsnew.CreateInfoExtractionRuleResponse, error) {
	rule, err := makeInfoExtractionRuleModel(req.GetNewInfoExtractionRule())
	if err != nil {
		return &whatsnew.CreateInfoExtractionRuleResponse{Errors: s.makeErrors(req, err)}, nil
	}

	ret := s.db.Create(rule)
	if ret.Error != nil {
		return &whatsnew.CreateInfoExtractionRuleResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.CreateInfoExtractionRuleResponse{
		Data: &whatsnew.CreateInfoExtractionRuleData{
			InfoExtractionRuleId: fmt.Sprintf("%d", rule.ID),
		},
	}
	return resp, nil
}

// GetInfoExtractionRule gets a InfoExtractionRule.
func (s *Server) GetInfoExtractionRule(
	_ context.Context,
	req *whatsnew.GetInfoExtractionRuleRequest,
) (*whatsnew.GetInfoExtractionRuleResponse, error) {
	var rule models.InfoExtractionRule
	ret := s.db.First(&rule, "id = ?", req.GetId())
	if ret.Error != nil {
		return &whatsnew.GetInfoExtractionRuleResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.GetInfoExtractionRuleResponse{
		Data: &whatsnew.GetInfoExtractionRuleData{
			InfoExtractionRule: makeAPIInfoExtractionRule(rule),
		},
	}
	return resp, nil
}

// UpdateInfoExtractionRule updates a InfoExtractionRule.
func (s *Server) UpdateInfoExtractionRule(
	ctx context.Context,
	req *whatsnew.UpdateInfoExtractionRuleRequest,
) (*whatsnew.UpdateInfoExtractionRuleResponse, error) {
	var rule models.InfoExtractionRule

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&rule, "id = ?", req.GetId())
		if ret.Error != nil {
			return ret.Error
		}

		ur := req.GetUpdatedInfoExtractionRule()

		arExpr := ur.GetAnswerRegexp()
		ar, err := regexp.Compile(arExpr)
		if err != nil {
			return fmt.Errorf("invalid AnswerRegexp value %#v: %v", arExpr, err)
		}

		rule.Label = ur.GetLabel()
		rule.Question = ur.GetQuestion()
		rule.AnswerRegexp.Regexp = ar
		rule.Threshold = ur.GetThreshold()
		rule.Enabled = ur.GetEnabled()

		ret = tx.Save(&rule)
		return ret.Error
	})

	if err != nil {
		return &whatsnew.UpdateInfoExtractionRuleResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.UpdateInfoExtractionRuleResponse{
		Data: &whatsnew.UpdateInfoExtractionRuleData{
			InfoExtractionRule: makeAPIInfoExtractionRule(rule),
		},
	}
	return resp, nil
}

// DeleteInfoExtractionRule deletes a InfoExtractionRule.
func (s *Server) DeleteInfoExtractionRule(
	ctx context.Context,
	req *whatsnew.DeleteInfoExtractionRuleRequest,
) (*whatsnew.DeleteInfoExtractionRuleResponse, error) {
	var rule models.InfoExtractionRule

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&rule, "id = ?", req.GetId())
		if ret.Error != nil {
			return ret.Error
		}

		var extractedInfosCount int64
		ret = tx.Model(&models.ExtractedInfo{}).
			Where("info_extraction_rule_id = ?", rule.ID).
			Limit(1).Count(&extractedInfosCount)
		if ret.Error != nil {
			return ret.Error
		}

		if extractedInfosCount == 0 {
			ret = tx.Unscoped().Delete(&rule)
		} else {
			ret = tx.Delete(&rule)
		}
		return ret.Error
	})

	if err != nil {
		return &whatsnew.DeleteInfoExtractionRuleResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.DeleteInfoExtractionRuleResponse{
		Data: &whatsnew.DeleteInfoExtractionRuleData{
			DeletedInfoExtractionRuleId: fmt.Sprintf("%d", rule.ID),
		},
	}
	return resp, nil
}

func makeInfoExtractionRuleModel(reqRule *whatsnew.NewInfoExtractionRule) (*models.InfoExtractionRule, error) {
	arExpr := reqRule.GetAnswerRegexp()
	ar, err := regexp.Compile(arExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid AnswerRegexp value %#v: %v", arExpr, err)
	}

	return &models.InfoExtractionRule{
		Label:        reqRule.GetLabel(),
		Question:     reqRule.GetQuestion(),
		AnswerRegexp: types.Regexp{Regexp: ar},
		Threshold:    reqRule.GetThreshold(),
		Enabled:      reqRule.GetEnabled(),
	}, nil
}
