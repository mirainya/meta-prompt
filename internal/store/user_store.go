package store

import (
	"meta-prompt/internal/model"

	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *UserStore) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := s.db.First(&user, id).Error
	return &user, err
}

func (s *UserStore) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (s *UserStore) DeductCredit(userID int64, amount int) error {
	result := s.db.Model(&model.User{}).
		Where("id = ? AND credits >= ?", userID, amount).
		Update("credits", gorm.Expr("credits - ?", amount))
	if result.RowsAffected == 0 {
		return ErrInsufficientCredits
	}
	return result.Error
}

func (s *UserStore) AddCredits(userID int64, amount int) error {
	return s.db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("credits", gorm.Expr("credits + ?", amount)).Error
}

func (s *UserStore) ListAll(limit, offset int) ([]model.User, error) {
	var users []model.User
	err := s.db.Order("id").Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (s *UserStore) CountAll() (int64, error) {
	var count int64
	err := s.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

func (s *UserStore) SetCredits(userID int64, amount int) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("credits", amount).Error
}

func (s *UserStore) SetDisabled(userID int64, disabled bool) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("disabled", disabled).Error
}

func (s *UserStore) UpdatePassword(userID int64, hash string) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", hash).Error
}

func (s *UserStore) SetRole(userID int64, role string) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error
}

func (s *UserStore) SetAllowedModels(userID int64, models []string) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("allowed_models", model.StringArray(models)).Error
}
