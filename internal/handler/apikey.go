package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/model"
	"meta-prompt/internal/store"
)

type APIKeyHandler struct {
	store        *store.APIKeyStore
	historyStore *store.HistoryStore
}

func NewAPIKeyHandler(s *store.APIKeyStore, hs ...*store.HistoryStore) *APIKeyHandler {
	h := &APIKeyHandler{store: s}
	if len(hs) > 0 {
		h.historyStore = hs[0]
	}
	return h
}

type CreateKeyRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

// List 用户查看自己的 API Keys
func (h *APIKeyHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	keys, err := h.store.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

// Create 用户创建 API Key，明文只返回一次
func (h *APIKeyHandler) Create(c *gin.Context) {
	var req CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")
	rawKey := generateKey()
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawKey)))

	key := &model.APIKey{
		KeyHash:  hash,
		Prefix:   rawKey[:8],
		Name:     req.Name,
		UserID:   userID,
		IsActive: true,
	}

	if err := h.store.Create(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      key.ID,
		"name":    key.Name,
		"key":     rawKey, // 明文只展示这一次
		"prefix":  key.Prefix,
		"created": key.CreatedAt,
	})
}

// Revoke 用户撤销自己的 Key
func (h *APIKeyHandler) Revoke(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.store.Deactivate(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

// AdminCreate Admin 为指定用户创建 Key
func (h *APIKeyHandler) AdminCreate(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		UserID    int64  `json:"user_id" binding:"required"`
		RateLimit int    `json:"rate_limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawKey := generateKey()
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawKey)))

	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 60
	}

	key := &model.APIKey{
		KeyHash:   hash,
		Prefix:    rawKey[:8],
		Name:      req.Name,
		UserID:    req.UserID,
		RateLimit: rateLimit,
		IsActive:  true,
	}

	if err := h.store.Create(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":     key.ID,
		"name":   key.Name,
		"key":    rawKey,
		"prefix": key.Prefix,
		"user_id": req.UserID,
	})
}

// AdminList Admin 查看所有 Key
func (h *APIKeyHandler) AdminList(c *gin.Context) {
	keys, err := h.store.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list keys"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

// AdminRevoke Admin 撤销任意 Key
func (h *APIKeyHandler) AdminRevoke(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.store.DeactivateByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "revoked"})
}

func generateKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "mp_" + hex.EncodeToString(b)
}

// Stats 用户查看自己某个 Key 的使用统计
func (h *APIKeyHandler) Stats(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetInt64("user_id")
	keys, _ := h.store.ListByUser(userID)
	var found *model.APIKey
	for i := range keys {
		if keys[i].ID == id {
			found = &keys[i]
			break
		}
	}
	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	var totalCalls int64
	if h.historyStore != nil {
		totalCalls = h.historyStore.CountBySource(userID, "api")
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            found.ID,
		"name":          found.Name,
		"prefix":        found.Prefix,
		"rate_limit":    found.RateLimit,
		"credits_quota": found.CreditsQuota,
		"is_active":     found.IsActive,
		"last_used_at":  found.LastUsedAt,
		"total_calls":   totalCalls,
	})
}
