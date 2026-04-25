package store

import (
	"time"

	"meta-prompt/internal/model"

	"gorm.io/gorm"
)

type HistoryStore struct {
	db *gorm.DB
}

func NewHistoryStore(db *gorm.DB) *HistoryStore {
	return &HistoryStore{db: db}
}

func (s *HistoryStore) Create(h *model.History) error {
	return s.db.Create(h).Error
}

func (s *HistoryStore) GetByID(id int64) (*model.History, error) {
	var h model.History
	err := s.db.First(&h, id).Error
	return &h, err
}

func (s *HistoryStore) List(limit, offset int) ([]model.History, error) {
	var list []model.History
	err := s.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, err
}

func (s *HistoryStore) ListByUser(userID int64, limit, offset int) ([]model.History, error) {
	var list []model.History
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&list).Error
	return list, err
}

func (s *HistoryStore) CountAll() (int64, error) {
	var count int64
	err := s.db.Model(&model.History{}).Count(&count).Error
	return count, err
}

func (s *HistoryStore) CountToday() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := s.db.Model(&model.History{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}

// UpdateStep 更新推演进度和中间结果
func (s *HistoryStore) UpdateStep(id int64, step int, fields map[string]any) error {
	fields["current_step"] = step
	fields["updated_at"] = time.Now()
	return s.db.Model(&model.History{}).Where("id = ?", id).Updates(fields).Error
}

// Finish 标记推演完成
func (s *HistoryStore) Finish(id int64, fields map[string]any) error {
	fields["status"] = "done"
	fields["updated_at"] = time.Now()
	return s.db.Model(&model.History{}).Where("id = ?", id).Updates(fields).Error
}

// Fail 标记推演失败
func (s *HistoryStore) Fail(id int64, errMsg string) error {
	return s.db.Model(&model.History{}).Where("id = ?", id).Updates(map[string]any{
		"status":     "failed",
		"error_msg":  errMsg,
		"updated_at": time.Now(),
	}).Error
}

// GetRunningByUser 查询用户正在进行的推演
func (s *HistoryStore) GetRunningByUser(userID int64) (*model.History, error) {
	var h model.History
	err := s.db.Where("user_id = ? AND status = ?", userID, "running").Order("created_at DESC").First(&h).Error
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// Cancel 取消正在进行的推演（仅限本人 + running 状态）
func (s *HistoryStore) Cancel(id int64, userID int64) error {
	result := s.db.Model(&model.History{}).
		Where("id = ? AND user_id = ? AND status = ?", id, userID, "running").
		Updates(map[string]any{
			"status":     "cancelled",
			"error_msg":  "用户取消",
			"updated_at": time.Now(),
		})
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return result.Error
}
