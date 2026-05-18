package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/model"
	"meta-prompt/internal/store"
)

type ChannelHandler struct {
	channelStore *store.ChannelStore
	userStore    *store.UserStore
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func NewChannelHandler(cs *store.ChannelStore, us *store.UserStore) *ChannelHandler {
	return &ChannelHandler{channelStore: cs, userStore: us}
}

// === Source CRUD ===

func (h *ChannelHandler) ListSources(c *gin.Context) {
	sources, err := h.channelStore.ListSources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// mask api keys
	for i := range sources {
		sources[i].APIKey = maskAPIKey(sources[i].APIKey)
	}
	c.JSON(http.StatusOK, sources)
}

func (h *ChannelHandler) CreateSource(c *gin.Context) {
	var src model.ChannelSource
	if err := c.ShouldBindJSON(&src); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.channelStore.CreateSource(&src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, src)
}

func (h *ChannelHandler) UpdateSource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.channelStore.GetSource(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var input model.ChannelSource
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing.Name = input.Name
	existing.BaseURL = input.BaseURL
	if input.APIKey != "" {
		existing.APIKey = input.APIKey
	}
	existing.ProxyURL = input.ProxyURL
	existing.Enabled = input.Enabled
	if err := h.channelStore.UpdateSource(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *ChannelHandler) DeleteSource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.channelStore.DeleteSource(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// === Sync ===

func (h *ChannelHandler) SyncSource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	src, err := h.channelStore.GetSource(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
		return
	}

	models, err := fetchRemoteModels(src)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("sync failed: %v", err)})
		return
	}

	if err := h.channelStore.SyncModels(id, models); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chatCount := 0
	for _, m := range models {
		if m.ModelType == "chat" {
			chatCount++
		}
	}
	c.JSON(http.StatusOK, gin.H{"synced": chatCount, "total": len(models)})
}

func fetchRemoteModels(src *model.ChannelSource) ([]model.ChannelModel, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	if src.ProxyURL != "" {
		if u, err := url.Parse(src.ProxyURL); err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
		}
	}

	apiURL := strings.TrimRight(src.BaseURL, "/") + "/v1/models"
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+src.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID      string `json:"id"`
			Type    string `json:"type"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	now := time.Now()
	models := make([]model.ChannelModel, 0, len(result.Data))
	for _, m := range result.Data {
		modelType := m.Type
		if modelType == "" {
			modelType = classifyModel(m.ID)
		}
		models = append(models, model.ChannelModel{
			ModelCode: m.ID,
			ModelType: modelType,
			SyncedAt:  now,
		})
	}
	return models, nil
}

func classifyModel(id string) string {
	lower := strings.ToLower(id)
	for _, kw := range []string{"dall-e", "tts", "whisper", "audio", "img", "image", "生图", "veo", "sora", "banana", "video"} {
		if strings.Contains(lower, kw) {
			return "media"
		}
	}
	for _, kw := range []string{"embedding", "embed"} {
		if strings.Contains(lower, kw) {
			return "embedding"
		}
	}
	return "chat"
}

// === Model Management ===

func (h *ChannelHandler) ListModels(c *gin.Context) {
	models, err := h.channelStore.ListModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type resp struct {
		ID               int64  `json:"id"`
		SourceID         int64  `json:"source_id"`
		SourceName       string `json:"source_name"`
		Code             string `json:"code"`
		Type             string `json:"type"`
		BillingType      string `json:"billing_type"`
		CreditsPerCall   int    `json:"credits_per_call"`
		InputTokenPrice  int    `json:"input_token_price"`
		OutputTokenPrice int    `json:"output_token_price"`
		Enabled          bool   `json:"enabled"`
		IsDefault        bool   `json:"is_default"`
	}
	out := make([]resp, 0, len(models))
	for _, m := range models {
		if m.ModelType != "chat" {
			continue
		}
		out = append(out, resp{
			ID: m.ID, SourceID: m.SourceID, SourceName: m.Source.Name,
			Code: m.ModelCode, Type: m.ModelType, BillingType: m.BillingType,
			CreditsPerCall: m.CreditsPerCall, InputTokenPrice: m.InputTokenPrice,
			OutputTokenPrice: m.OutputTokenPrice, Enabled: m.Enabled, IsDefault: m.IsDefault,
		})
	}
	c.JSON(http.StatusOK, out)
}

type UpdateModelRequest struct {
	Enabled          *bool   `json:"enabled"`
	CreditsPerCall   *int    `json:"credits_per_call"`
	BillingType      *string `json:"billing_type"`
	InputTokenPrice  *int    `json:"input_token_price"`
	OutputTokenPrice *int    `json:"output_token_price"`
	IsDefault        *bool   `json:"is_default"`
}

func (h *ChannelHandler) UpdateModel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input UpdateModelRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var m model.ChannelModel
	m.ID = id
	// 先查出来
	models, _ := h.channelStore.ListModels()
	found := false
	for _, mod := range models {
		if mod.ID == id {
			m = mod
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	if input.Enabled != nil {
		m.Enabled = *input.Enabled
	}
	if input.CreditsPerCall != nil {
		m.CreditsPerCall = *input.CreditsPerCall
	}
	if input.BillingType != nil {
		m.BillingType = *input.BillingType
	}
	if input.InputTokenPrice != nil {
		m.InputTokenPrice = *input.InputTokenPrice
	}
	if input.OutputTokenPrice != nil {
		m.OutputTokenPrice = *input.OutputTokenPrice
	}
	if input.IsDefault != nil && *input.IsDefault {
		h.channelStore.SetDefaultModel(m.ID)
		m.IsDefault = true
	}

	if err := h.channelStore.UpdateModel(&m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

// === Public: list available models for users ===

func (h *ChannelHandler) ListAvailableModels(c *gin.Context) {
	models, err := h.channelStore.ListEnabledChatModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 按用户 allowed_models 过滤
	userID := c.GetInt64("user_id")
	if userID > 0 {
		if u, err := h.userStore.GetByID(userID); err == nil && len(u.AllowedModels) > 0 {
			allowed := make(map[string]bool, len(u.AllowedModels))
			for _, code := range u.AllowedModels {
				allowed[code] = true
			}
			filtered := models[:0]
			for _, m := range models {
				if allowed[m.ModelCode] {
					filtered = append(filtered, m)
				}
			}
			models = filtered
		}
	}

	type modelInfo struct {
		Code             string `json:"code"`
		Type             string `json:"type"`
		BillingType      string `json:"billing_type"`
		CreditsPerCall   int    `json:"credits_per_call"`
		InputTokenPrice  int    `json:"input_token_price"`
		OutputTokenPrice int    `json:"output_token_price"`
		IsDefault        bool   `json:"is_default"`
	}
	resp := make([]modelInfo, len(models))
	for i, m := range models {
		resp[i] = modelInfo{
			Code: m.ModelCode, Type: m.ModelType, BillingType: m.BillingType,
			CreditsPerCall: m.CreditsPerCall, InputTokenPrice: m.InputTokenPrice,
			OutputTokenPrice: m.OutputTokenPrice, IsDefault: m.IsDefault,
		}
	}
	c.JSON(http.StatusOK, resp)
}
