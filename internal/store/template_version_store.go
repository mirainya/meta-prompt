package store

import (
	"meta-prompt/internal/model"

	"gorm.io/gorm"
)

type TemplateVersionStore struct {
	db *gorm.DB
}

func NewTemplateVersionStore(db *gorm.DB) *TemplateVersionStore {
	return &TemplateVersionStore{db: db}
}

func (s *TemplateVersionStore) Save(templateID int64, prompt string, version int) error {
	v := &model.TemplateVersion{
		TemplateID: templateID,
		Prompt:     prompt,
		Version:    version,
	}
	return s.db.Create(v).Error
}

func (s *TemplateVersionStore) ListByTemplate(templateID int64) ([]model.TemplateVersion, error) {
	var versions []model.TemplateVersion
	err := s.db.Where("template_id = ?", templateID).Order("version DESC").Find(&versions).Error
	return versions, err
}

func (s *TemplateVersionStore) GetByVersion(templateID int64, version int) (*model.TemplateVersion, error) {
	var v model.TemplateVersion
	err := s.db.Where("template_id = ? AND version = ?", templateID, version).First(&v).Error
	return &v, err
}
