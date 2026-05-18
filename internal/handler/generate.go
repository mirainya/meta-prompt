package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/service"
	"meta-prompt/internal/store"
)

type GenerateHandler struct {
	pipeline     *service.Pipeline
	userStore    *store.UserStore
	historyStore *store.HistoryStore
	channelStore *store.ChannelStore
	defaultModel string
}

func NewGenerateHandler(p *service.Pipeline, defaultModel string, us *store.UserStore, hs *store.HistoryStore, cs *store.ChannelStore) *GenerateHandler {
	return &GenerateHandler{pipeline: p, defaultModel: defaultModel, userStore: us, historyStore: hs, channelStore: cs}
}

type GenerateRequest struct {
	Input               string `json:"input" binding:"required"`
	Model               string `json:"model"`
	AnalyzerTemplateID  *int64 `json:"analyzer_template_id"`
	ArchitectTemplateID *int64 `json:"architect_template_id"`
	WriterTemplateID    *int64 `json:"writer_template_id"`
	ReviewerTemplateID  *int64 `json:"reviewer_template_id"`
}

func (h *GenerateHandler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model == "" {
		if dm, err := h.channelStore.GetDefaultModel(); err == nil {
			req.Model = dm.ModelCode
		} else {
			req.Model = h.defaultModel
		}
	}

	// 查模型定价
	cm, err := h.channelStore.GetModelByCode(req.Model)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model not available: " + req.Model})
		return
	}
	if !cm.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model disabled: " + req.Model})
		return
	}

	userID := c.GetInt64("user_id")
	credits := cm.CreditsPerCall

	if err := h.userStore.DeductCredit(userID, credits); err != nil {
		if errors.Is(err, store.ErrInsufficientCredits) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient credits"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	historyID, err := h.pipeline.ExecuteAsync(service.PipelineRequest{
		Input:               req.Input,
		UserID:              userID,
		Model:               req.Model,
		AnalyzerTemplateID:  req.AnalyzerTemplateID,
		ArchitectTemplateID: req.ArchitectTemplateID,
		WriterTemplateID:    req.WriterTemplateID,
		ReviewerTemplateID:  req.ReviewerTemplateID,
	})
	if err != nil {
		h.userStore.AddCredits(userID, credits)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": historyID})
}

func (h *GenerateHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID := c.GetInt64("user_id")

	if err := h.historyStore.Cancel(id, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found or not running"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.pipeline.Cancel(id)
	_ = h.userStore.AddCredits(userID, 1)

	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}
