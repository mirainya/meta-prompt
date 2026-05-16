package store

import (
	"crypto/sha256"
	"fmt"
	"time"

	"meta-prompt/internal/model"

	"gorm.io/gorm"
)

type APIKeyStore struct {
	db *gorm.DB
}

func NewAPIKeyStore(db *gorm.DB) *APIKeyStore {
	return &APIKeyStore{db: db}
}

func (s *APIKeyStore) Create(key *model.APIKey) error {
	return s.db.Create(key).Error
}

func (s *APIKeyStore) GetByHash(hash string) (*model.APIKey, error) {
	var k model.APIKey
	err := s.db.Where("key_hash = ? AND is_active = ?", hash, true).First(&k).Error
	return &k, err
}

func (s *APIKeyStore) ValidateKey(rawKey string) (*model.APIKey, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawKey)))
	k, err := s.GetByHash(hash)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	s.db.Model(k).Update("last_used_at", &now)
	return k, nil
}

func (s *APIKeyStore) ListByUser(userID int64) ([]model.APIKey, error) {
	var keys []model.APIKey
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (s *APIKeyStore) Deactivate(id int64, userID int64) error {
	return s.db.Model(&model.APIKey{}).Where("id = ? AND user_id = ?", id, userID).Update("is_active", false).Error
}

func (s *APIKeyStore) DeactivateByID(id int64) error {
	return s.db.Model(&model.APIKey{}).Where("id = ?", id).Update("is_active", false).Error
}

func (s *APIKeyStore) ListAll() ([]model.APIKey, error) {
	var keys []model.APIKey
	err := s.db.Order("created_at DESC").Find(&keys).Error
	return keys, err
}
