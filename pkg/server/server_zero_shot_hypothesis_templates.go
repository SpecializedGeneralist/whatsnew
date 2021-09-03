// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server/whatsnew"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetZeroShotHypothesisTemplates gets all ZeroShotHypothesisTemplates with
// their related ZeroShotHypothesisLabels.
func (s *Server) GetZeroShotHypothesisTemplates(
	_ context.Context,
	req *whatsnew.GetZeroShotHypothesisTemplatesRequest,
) (*whatsnew.GetZeroShotHypothesisTemplatesResponse, error) {
	query := s.db.Preload("Labels").Order("id")
	if len(req.GetAfter()) > 0 {
		query = query.Where("id > ?", req.GetAfter())
	}
	if req.GetFirst() > 0 {
		query = query.Limit(int(req.GetFirst()))
	}

	var templates []models.ZeroShotHypothesisTemplate
	ret := query.Find(&templates)
	if ret.Error != nil {
		return &whatsnew.GetZeroShotHypothesisTemplatesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	respTemplates := make([]*whatsnew.ZeroShotHypothesisTemplate, len(templates))
	for i, template := range templates {
		respTemplates[i] = makeAPIZeroShotHypothesisTemplate(template)
	}

	resp := &whatsnew.GetZeroShotHypothesisTemplatesResponse{
		Data: &whatsnew.GetZeroShotHypothesisTemplatesData{
			ZeroShotHypothesisTemplates: respTemplates,
		},
	}
	return resp, nil
}

// CreateZeroShotHypothesisTemplates creates new ZeroShotHypothesisTemplates
// with related ZeroShotHypothesisLabels.
func (s *Server) CreateZeroShotHypothesisTemplates(
	_ context.Context,
	req *whatsnew.CreateZeroShotHypothesisTemplatesRequest,
) (*whatsnew.CreateZeroShotHypothesisTemplatesResponse, error) {
	reqTemplates := req.GetNewZeroShotHypothesisTemplates().GetZeroShotHypothesisTemplates()

	templates := make([]models.ZeroShotHypothesisTemplate, len(reqTemplates))
	for i, reqTemplate := range reqTemplates {
		templates[i] = makeZeroShotHypothesisTemplateModel(reqTemplate)
	}

	ret := s.db.Create(&templates)
	if ret.Error != nil {
		return &whatsnew.CreateZeroShotHypothesisTemplatesResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}

	ids := make([]string, len(templates))
	for i, template := range templates {
		ids[i] = fmt.Sprintf("%d", template.ID)
	}

	resp := &whatsnew.CreateZeroShotHypothesisTemplatesResponse{
		Data: &whatsnew.CreateZeroShotHypothesisTemplatesData{
			ZeroShotHypothesisTemplateIds: ids,
		},
	}
	return resp, nil
}

// CreateZeroShotHypothesisTemplate creates a new ZeroShotHypothesisTemplate
// with related ZeroShotHypothesisLabels.
func (s *Server) CreateZeroShotHypothesisTemplate(
	_ context.Context,
	req *whatsnew.CreateZeroShotHypothesisTemplateRequest,
) (*whatsnew.CreateZeroShotHypothesisTemplateResponse, error) {
	template := makeZeroShotHypothesisTemplateModel(req.GetNewZeroShotHypothesisTemplate())
	ret := s.db.Create(&template)
	if ret.Error != nil {
		return &whatsnew.CreateZeroShotHypothesisTemplateResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.CreateZeroShotHypothesisTemplateResponse{
		Data: &whatsnew.CreateZeroShotHypothesisTemplateData{
			ZeroShotHypothesisTemplateId: fmt.Sprintf("%d", template.ID),
		},
	}
	return resp, nil
}

// GetZeroShotHypothesisTemplate gets a ZeroShotHypothesisTemplate with
// its related ZeroShotHypothesisLabels.
func (s *Server) GetZeroShotHypothesisTemplate(
	_ context.Context,
	req *whatsnew.GetZeroShotHypothesisTemplateRequest,
) (*whatsnew.GetZeroShotHypothesisTemplateResponse, error) {
	var template models.ZeroShotHypothesisTemplate
	ret := s.db.Preload("Labels").First(&template, "id = ?", req.GetId())
	if ret.Error != nil {
		return &whatsnew.GetZeroShotHypothesisTemplateResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.GetZeroShotHypothesisTemplateResponse{
		Data: &whatsnew.GetZeroShotHypothesisTemplateData{
			ZeroShotHypothesisTemplate: makeAPIZeroShotHypothesisTemplate(template),
		},
	}
	return resp, nil
}

// UpdateZeroShotHypothesisTemplate updates a ZeroShotHypothesisTemplate.
func (s *Server) UpdateZeroShotHypothesisTemplate(
	ctx context.Context,
	req *whatsnew.UpdateZeroShotHypothesisTemplateRequest,
) (*whatsnew.UpdateZeroShotHypothesisTemplateResponse, error) {
	var template models.ZeroShotHypothesisTemplate

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Labels").First(&template, "id = ?", req.GetId())
		if ret.Error != nil {
			return ret.Error
		}

		ut := req.GetUpdatedZeroShotHypothesisTemplate()

		template.Enabled = ut.GetEnabled()
		template.Text = ut.GetText()
		template.MultiClass = ut.GetMultiClass()

		ret = tx.Save(&template)
		return ret.Error
	})

	if err != nil {
		return &whatsnew.UpdateZeroShotHypothesisTemplateResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.UpdateZeroShotHypothesisTemplateResponse{
		Data: &whatsnew.UpdateZeroShotHypothesisTemplateData{
			ZeroShotHypothesisTemplate: makeAPIZeroShotHypothesisTemplate(template),
		},
	}
	return resp, nil
}

// DeleteZeroShotHypothesisTemplate deletes a ZeroShotHypothesisTemplate.
func (s *Server) DeleteZeroShotHypothesisTemplate(
	ctx context.Context,
	req *whatsnew.DeleteZeroShotHypothesisTemplateRequest,
) (*whatsnew.DeleteZeroShotHypothesisTemplateResponse, error) {
	var template models.ZeroShotHypothesisTemplate

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&template, "id = ?", req.GetId())
		if ret.Error != nil {
			return ret.Error
		}

		var classesCount int64
		ret = tx.Model(&models.ZeroShotClass{}).
			Where("zero_shot_hypothesis_template_id = ?", template.ID).
			Limit(1).Count(&classesCount)
		if ret.Error != nil {
			return ret.Error
		}

		if classesCount == 0 {
			ret = tx.Unscoped().Delete(&template)
		} else {
			ret = tx.Delete(&template)
		}
		return ret.Error
	})

	if err != nil {
		return &whatsnew.DeleteZeroShotHypothesisTemplateResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.DeleteZeroShotHypothesisTemplateResponse{
		Data: &whatsnew.DeleteZeroShotHypothesisTemplateData{
			DeletedZeroShotHypothesisTemplateId: fmt.Sprintf("%d", template.ID),
		},
	}
	return resp, nil
}

// CreateZeroShotHypothesisLabels creates new ZeroShotHypothesisLabels.
func (s *Server) CreateZeroShotHypothesisLabels(
	ctx context.Context,
	req *whatsnew.CreateZeroShotHypothesisLabelsRequest,
) (*whatsnew.CreateZeroShotHypothesisLabelsResponse, error) {
	var ids []string

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var template models.ZeroShotHypothesisTemplate
		ret := tx.First(&template, "id = ?", req.GetTemplateId())
		if ret.Error != nil {
			return ret.Error
		}

		reqLabels := req.GetNewZeroShotHypothesisLabels().GetZeroShotHypothesisLabels()

		labels := make([]models.ZeroShotHypothesisLabel, len(reqLabels))
		for i, reqLabel := range reqLabels {
			labels[i] = models.ZeroShotHypothesisLabel{
				ZeroShotHypothesisTemplateID: template.ID,
				Enabled:                      reqLabel.GetEnabled(),
				Text:                         reqLabel.GetText(),
			}
		}

		ret = s.db.Create(&labels)
		if ret.Error != nil {
			return ret.Error
		}

		ids = make([]string, len(labels))
		for i, template := range labels {
			ids[i] = fmt.Sprintf("%d", template.ID)
		}
		return nil
	})

	if err != nil {
		return &whatsnew.CreateZeroShotHypothesisLabelsResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.CreateZeroShotHypothesisLabelsResponse{
		Data: &whatsnew.CreateZeroShotHypothesisLabelsData{
			ZeroShotHypothesisLabelIds: ids,
		},
	}
	return resp, nil
}

// CreateZeroShotHypothesisLabel creates new ZeroShotHypothesisLabel.
func (s *Server) CreateZeroShotHypothesisLabel(
	ctx context.Context,
	req *whatsnew.CreateZeroShotHypothesisLabelRequest,
) (*whatsnew.CreateZeroShotHypothesisLabelResponse, error) {
	var label models.ZeroShotHypothesisLabel

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var template models.ZeroShotHypothesisTemplate
		ret := tx.First(&template, "id = ?", req.GetTemplateId())
		if ret.Error != nil {
			return ret.Error
		}

		reqLabel := req.GetNewZeroShotHypothesisLabel()
		label = models.ZeroShotHypothesisLabel{
			ZeroShotHypothesisTemplateID: template.ID,
			Enabled:                      reqLabel.GetEnabled(),
			Text:                         reqLabel.GetText(),
		}

		ret = s.db.Create(&label)
		if ret.Error != nil {
			return ret.Error
		}
		return nil
	})

	if err != nil {
		return &whatsnew.CreateZeroShotHypothesisLabelResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.CreateZeroShotHypothesisLabelResponse{
		Data: &whatsnew.CreateZeroShotHypothesisLabelData{
			ZeroShotHypothesisLabelId: fmt.Sprintf("%d", label.ID),
		},
	}
	return resp, nil
}

// GetZeroShotHypothesisLabel gets a ZeroShotHypothesisLabel.
func (s *Server) GetZeroShotHypothesisLabel(
	_ context.Context,
	req *whatsnew.GetZeroShotHypothesisLabelRequest,
) (*whatsnew.GetZeroShotHypothesisLabelResponse, error) {
	var label models.ZeroShotHypothesisLabel
	ret := s.db.First(&label, "id = ? AND zero_shot_hypothesis_template_id = ?", req.GetLabelId(), req.GetTemplateId())
	if ret.Error != nil {
		return &whatsnew.GetZeroShotHypothesisLabelResponse{Errors: s.makeErrors(req, ret.Error)}, nil
	}
	resp := &whatsnew.GetZeroShotHypothesisLabelResponse{
		Data: &whatsnew.GetZeroShotHypothesisLabelData{
			ZeroShotHypothesisLabel: makeAPIZeroShotHypothesisLabel(label),
		},
	}
	return resp, nil
}

// UpdateZeroShotHypothesisLabel updates a ZeroShotHypothesisLabel.
func (s *Server) UpdateZeroShotHypothesisLabel(
	ctx context.Context,
	req *whatsnew.UpdateZeroShotHypothesisLabelRequest,
) (*whatsnew.UpdateZeroShotHypothesisLabelResponse, error) {
	var label models.ZeroShotHypothesisLabel

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&label, "id = ? AND zero_shot_hypothesis_template_id = ?", req.GetLabelId(), req.GetTemplateId())
		if ret.Error != nil {
			return ret.Error
		}

		ul := req.GetUpdatedZeroShotHypothesisLabel()

		label.Enabled = ul.GetEnabled()
		label.Text = ul.GetText()

		ret = tx.Save(&label)
		return ret.Error
	})

	if err != nil {
		return &whatsnew.UpdateZeroShotHypothesisLabelResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.UpdateZeroShotHypothesisLabelResponse{
		Data: &whatsnew.UpdateZeroShotHypothesisLabelData{
			ZeroShotHypothesisLabel: makeAPIZeroShotHypothesisLabel(label),
		},
	}
	return resp, nil
}

// DeleteZeroShotHypothesisLabel deletes a ZeroShotHypothesisLabel.
func (s *Server) DeleteZeroShotHypothesisLabel(
	ctx context.Context,
	req *whatsnew.DeleteZeroShotHypothesisLabelRequest,
) (*whatsnew.DeleteZeroShotHypothesisLabelResponse, error) {
	var label models.ZeroShotHypothesisLabel

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ret := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&label, "id = ? AND zero_shot_hypothesis_template_id = ?", req.GetLabelId(), req.GetTemplateId())
		if ret.Error != nil {
			return ret.Error
		}

		var classesCount int64
		ret = tx.Model(&models.ZeroShotClass{}).
			Where("zero_shot_hypothesis_label_id = ?", label.ID).
			Limit(1).Count(&classesCount)
		if ret.Error != nil {
			return ret.Error
		}

		if classesCount == 0 {
			ret = tx.Unscoped().Delete(&label)
		} else {
			ret = tx.Delete(&label)
		}
		return ret.Error
	})

	if err != nil {
		return &whatsnew.DeleteZeroShotHypothesisLabelResponse{Errors: s.makeErrors(req, err)}, nil
	}

	resp := &whatsnew.DeleteZeroShotHypothesisLabelResponse{
		Data: &whatsnew.DeleteZeroShotHypothesisLabelData{
			DeletedZeroShotHypothesisLabelId: fmt.Sprintf("%d", label.ID),
		},
	}
	return resp, nil
}

func makeZeroShotHypothesisTemplateModel(
	reqTemplate *whatsnew.NewZeroShotHypothesisTemplate,
) models.ZeroShotHypothesisTemplate {
	reqLabels := reqTemplate.GetLabels()
	labels := make([]models.ZeroShotHypothesisLabel, len(reqLabels))
	for j, reqLabel := range reqLabels {
		labels[j] = models.ZeroShotHypothesisLabel{
			Enabled: reqLabel.GetEnabled(),
			Text:    reqLabel.GetText(),
		}
	}

	return models.ZeroShotHypothesisTemplate{
		Enabled:    reqTemplate.GetEnabled(),
		Text:       reqTemplate.GetText(),
		MultiClass: reqTemplate.GetMultiClass(),
		Labels:     labels,
	}
}
