package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"meta-prompt/internal/llm"
	"meta-prompt/internal/model"
	"meta-prompt/internal/store"
)

type AdminHandler struct {
	userStore      *store.UserStore
	llmConfigStore *store.LLMConfigStore
	historyStore   *store.HistoryStore
	providerMgr    *llm.ProviderManager
}

func NewAdminHandler(us *store.UserStore, lcs *store.LLMConfigStore, hs *store.HistoryStore, pm *llm.ProviderManager) *AdminHandler {
	return &AdminHandler{userStore: us, llmConfigStore: lcs, historyStore: hs, providerMgr: pm}
}

// ========== Dashboard ==========

func (h *AdminHandler) Dashboard(c *gin.Context) {
	userCount, _ := h.userStore.CountAll()
	totalGen, _ := h.historyStore.CountAll()
	todayGen, _ := h.historyStore.CountToday()

	c.JSON(http.StatusOK, gin.H{
		"user_count":        userCount,
		"total_generations": totalGen,
		"today_generations": todayGen,
	})
}

// ========== LLM Config ==========

type LLMConfigResponse struct {
	Provider  string `json:"provider"`
	Type      string `json:"type"`
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Enabled   bool   `json:"enabled"`
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func (h *AdminHandler) ListLLMConfigs(c *gin.Context) {
	configs, err := h.llmConfigStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := make([]LLMConfigResponse, len(configs))
	for i, cfg := range configs {
		resp[i] = LLMConfigResponse{
			Provider:  cfg.Provider,
			Type:      cfg.Type,
			APIKey:    maskAPIKey(cfg.APIKey),
			BaseURL:   cfg.BaseURL,
			Model:     cfg.Model,
			MaxTokens: cfg.MaxTokens,
			Enabled:   cfg.Enabled,
		}
	}
	c.JSON(http.StatusOK, resp)
}

type CreateLLMConfigRequest struct {
	Provider  string `json:"provider" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=openai_compatible claude gemini"`
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Model     string `json:"model" binding:"required"`
	MaxTokens int    `json:"max_tokens"`
	Enabled   bool   `json:"enabled"`
}

func (h *AdminHandler) CreateLLMConfig(c *gin.Context) {
	var req CreateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在
	if _, err := h.llmConfigStore.GetByProvider(req.Provider); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "provider already exists"})
		return
	}

	if req.MaxTokens <= 0 {
		req.MaxTokens = 4096
	}

	cfg := &model.LLMConfig{
		Provider:  req.Provider,
		Type:      req.Type,
		APIKey:    req.APIKey,
		BaseURL:   req.BaseURL,
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Enabled:   req.Enabled,
	}
	if err := h.llmConfigStore.Upsert(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rebuildSingleProvider(h.providerMgr, cfg)
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

type UpdateLLMConfigRequest struct {
	Type      *string `json:"type"`
	APIKey    *string `json:"api_key"`
	BaseURL   *string `json:"base_url"`
	Model     *string `json:"model"`
	MaxTokens *int    `json:"max_tokens"`
	Enabled   *bool   `json:"enabled"`
}

func (h *AdminHandler) UpdateLLMConfig(c *gin.Context) {
	provider := c.Param("provider")
	var req UpdateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := h.llmConfigStore.GetByProvider(provider)
	if err != nil {
		cfg = &model.LLMConfig{Provider: provider, Type: "openai_compatible"}
	}

	if req.Type != nil {
		cfg.Type = *req.Type
	}
	if req.APIKey != nil {
		cfg.APIKey = *req.APIKey
	}
	if req.BaseURL != nil {
		cfg.BaseURL = *req.BaseURL
	}
	if req.Model != nil {
		cfg.Model = *req.Model
	}
	if req.MaxTokens != nil {
		cfg.MaxTokens = *req.MaxTokens
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}

	if err := h.llmConfigStore.Upsert(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rebuildSingleProvider(h.providerMgr, cfg)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AdminHandler) DeleteLLMConfig(c *gin.Context) {
	provider := c.Param("provider")
	if err := h.llmConfigStore.Delete(provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.providerMgr.Remove(provider)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AdminHandler) TestLLMConfig(c *gin.Context) {
	provider := c.Param("provider")
	cfg, err := h.llmConfigStore.GetByProvider(provider)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
		return
	}

	p := buildProvider(cfg)
	if p == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot build provider, check api_key and type"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	resp, err := p.Chat(ctx, llm.ChatRequest{
		Messages:  []llm.Message{{Role: "user", Content: "Hi, reply with just 'ok'"}},
		MaxTokens: 10,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "reply": resp.Content})
}

// ========== Providers (public) ==========

type ProviderHandler struct {
	llmConfigStore *store.LLMConfigStore
}

func NewProviderHandler(lcs *store.LLMConfigStore) *ProviderHandler {
	return &ProviderHandler{llmConfigStore: lcs}
}

func (h *ProviderHandler) ListEnabled(c *gin.Context) {
	configs, err := h.llmConfigStore.ListEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type item struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	resp := make([]item, len(configs))
	for i, cfg := range configs {
		resp[i] = item{Provider: cfg.Provider, Model: cfg.Model}
	}
	c.JSON(http.StatusOK, resp)
}

// ========== User Management ==========

func (h *AdminHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	users, err := h.userStore.ListAll(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	total, _ := h.userStore.CountAll()
	c.JSON(http.StatusOK, gin.H{"users": users, "total": total})
}

type SetCreditsRequest struct {
	Credits int `json:"credits" binding:"min=0"`
}

func (h *AdminHandler) SetCredits(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req SetCreditsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userStore.SetCredits(id, req.Credits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AdminHandler) ToggleUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	user, err := h.userStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err := h.userStore.SetDisabled(id, !user.Disabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"disabled": !user.Disabled})
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.userStore.UpdatePassword(id, string(hash)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset"})
}

// ========== Provider Builder ==========

func buildProvider(cfg *model.LLMConfig) llm.Provider {
	if cfg.APIKey == "" {
		return nil
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	switch cfg.Type {
	case "claude":
		return llm.NewClaudeProviderWithBase(cfg.APIKey, baseURL, cfg.Model, cfg.MaxTokens)
	case "gemini":
		return llm.NewGeminiProviderWithBase(cfg.APIKey, baseURL, cfg.Model, cfg.MaxTokens)
	default: // openai_compatible
		return llm.NewOpenAIProviderWithBase(cfg.APIKey, baseURL, cfg.Model, cfg.MaxTokens)
	}
}

func rebuildSingleProvider(mgr *llm.ProviderManager, cfg *model.LLMConfig) {
	if !cfg.Enabled || cfg.APIKey == "" {
		mgr.Remove(cfg.Provider)
		return
	}
	p := buildProvider(cfg)
	if p != nil {
		mgr.Set(cfg.Provider, p)
	}
}

// RebuildAllProviders 从数据库重建所有 provider
func RebuildAllProviders(mgr *llm.ProviderManager, configs []model.LLMConfig) {
	for i := range configs {
		rebuildSingleProvider(mgr, &configs[i])
	}
}
