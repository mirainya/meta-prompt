package store

import (
	"meta-prompt/internal/model"

	"gorm.io/gorm"
)

type TemplateStore struct {
	db *gorm.DB
}

func NewTemplateStore(db *gorm.DB) *TemplateStore {
	return &TemplateStore{db: db}
}

func (s *TemplateStore) Create(t *model.Template) error {
	return s.db.Create(t).Error
}

func (s *TemplateStore) GetByID(id int64) (*model.Template, error) {
	var t model.Template
	err := s.db.First(&t, id).Error
	return &t, err
}

func (s *TemplateStore) GetDefault(stage string) (*model.Template, error) {
	var t model.Template
	err := s.db.Where("stage = ? AND is_default = ?", stage, true).First(&t).Error
	return &t, err
}

func (s *TemplateStore) List(stage string) ([]model.Template, error) {
	var list []model.Template
	q := s.db
	if stage != "" {
		q = q.Where("stage = ?", stage)
	}
	err := q.Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *TemplateStore) Update(t *model.Template) error {
	return s.db.Save(t).Error
}

func (s *TemplateStore) Delete(id int64) error {
	return s.db.Delete(&model.Template{}, id).Error
}
