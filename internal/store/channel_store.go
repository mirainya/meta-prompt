package store

import (
	"meta-prompt/internal/model"

	"gorm.io/gorm"
)

type ChannelStore struct {
	db *gorm.DB
}

func NewChannelStore(db *gorm.DB) *ChannelStore {
	return &ChannelStore{db: db}
}

// === ChannelSource ===

func (s *ChannelStore) ListSources() ([]model.ChannelSource, error) {
	var sources []model.ChannelSource
	err := s.db.Order("name").Find(&sources).Error
	return sources, err
}

func (s *ChannelStore) GetSource(id int64) (*model.ChannelSource, error) {
	var src model.ChannelSource
	err := s.db.First(&src, id).Error
	return &src, err
}

func (s *ChannelStore) CreateSource(src *model.ChannelSource) error {
	return s.db.Create(src).Error
}

func (s *ChannelStore) UpdateSource(src *model.ChannelSource) error {
	return s.db.Save(src).Error
}

func (s *ChannelStore) DeleteSource(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("source_id = ?", id).Delete(&model.ChannelModel{})
		return tx.Delete(&model.ChannelSource{}, id).Error
	})
}

// === ChannelModel ===

func (s *ChannelStore) ListModels() ([]model.ChannelModel, error) {
	var models []model.ChannelModel
	err := s.db.Preload("Source").Order("model_code").Find(&models).Error
	return models, err
}

func (s *ChannelStore) ListEnabledChatModels() ([]model.ChannelModel, error) {
	var models []model.ChannelModel
	err := s.db.Preload("Source").
		Joins("JOIN channel_sources ON channel_sources.id = channel_models.source_id").
		Where("channel_models.enabled = ? AND channel_models.model_type = ? AND channel_sources.enabled = ?", true, "chat", true).
		Order("model_code").Find(&models).Error
	return models, err
}

func (s *ChannelStore) GetModelByCode(code string) (*model.ChannelModel, error) {
	var m model.ChannelModel
	err := s.db.Preload("Source").Where("model_code = ?", code).First(&m).Error
	return &m, err
}

func (s *ChannelStore) UpdateModel(m *model.ChannelModel) error {
	return s.db.Save(m).Error
}

func (s *ChannelStore) SetDefaultModel(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Model(&model.ChannelModel{}).Where("is_default = ?", true).Update("is_default", false)
		return tx.Model(&model.ChannelModel{}).Where("id = ?", id).Update("is_default", true).Error
	})
}

func (s *ChannelStore) GetDefaultModel() (*model.ChannelModel, error) {
	var m model.ChannelModel
	err := s.db.Preload("Source").Where("is_default = ? AND enabled = ?", true, true).First(&m).Error
	return &m, err
}

func (s *ChannelStore) SyncModels(sourceID int64, models []model.ChannelModel) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取现有模型
		var existing []model.ChannelModel
		tx.Where("source_id = ?", sourceID).Find(&existing)
		existingMap := make(map[string]*model.ChannelModel, len(existing))
		for i := range existing {
			existingMap[existing[i].ModelCode] = &existing[i]
		}

		// 新增或更新
		syncedCodes := make(map[string]bool)
		for i := range models {
			models[i].SourceID = sourceID
			syncedCodes[models[i].ModelCode] = true
			if ex, ok := existingMap[models[i].ModelCode]; ok {
				// 保留用户设置的 enabled 和 credits_per_call
				tx.Model(ex).Updates(map[string]any{
					"model_type": models[i].ModelType,
					"synced_at":  models[i].SyncedAt,
				})
			} else {
				models[i].Enabled = true
				tx.Create(&models[i])
			}
		}

		// 删除远端已不存在的模型
		for code, ex := range existingMap {
			if !syncedCodes[code] {
				tx.Delete(ex)
			}
		}
		return nil
	})
}
