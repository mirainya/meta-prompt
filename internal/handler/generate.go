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
	pipeline        *service.Pipeline
	defaultProvider string
	userStore       *store.UserStore
	historyStore    *store.HistoryStore
}

func NewGenerateHandler(p *service.Pipeline, defaultProvider string, us *store.UserStore, hs *store.HistoryStore) *GenerateHandler {
	return &GenerateHandler{pipeline: p, defaultProvider: defaultProvider, userStore: us, historyStore: hs}
}

type GenerateRequest struct {
	Input               string `json:"input" binding:"required"`
	LLMProvider         string `json:"llm_provider"`
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

	if req.LLMProvider == "" {
		req.LLMProvider = h.defaultProvider
	}

	userID := c.GetInt64("user_id")

	// 扣积分
	if err := h.userStore.DeductCredit(userID, 1); err != nil {
		if errors.Is(err, store.ErrInsufficientCredits) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient credits"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 异步执行，立即返回 history id
	historyID, err := h.pipeline.ExecuteAsync(service.PipelineRequest{
		Input:               req.Input,
		UserID:              userID,
		LLMProvider:         req.LLMProvider,
		AnalyzerTemplateID:  req.AnalyzerTemplateID,
		ArchitectTemplateID: req.ArchitectTemplateID,
		WriterTemplateID:    req.WriterTemplateID,
		ReviewerTemplateID:  req.ReviewerTemplateID,
	})
	if err != nil {
		// 启动失败，退还积分
		h.userStore.AddCredits(userID, 1)
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

	// 标记数据库状态为 cancelled
	if err := h.historyStore.Cancel(id, userID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found or not running"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 取消后台 goroutine 的 context
	h.pipeline.Cancel(id)

	// 退还积分
	_ = h.userStore.AddCredits(userID, 1)

	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}
