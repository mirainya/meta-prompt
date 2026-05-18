package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/service"
	"meta-prompt/internal/store"
)

type OpenHandler struct {
	pipeline     *service.Pipeline
	historyStore *store.HistoryStore
	userStore    *store.UserStore
	channelStore *store.ChannelStore
	defaultModel string
}

func NewOpenHandler(p *service.Pipeline, hs *store.HistoryStore, us *store.UserStore, cs *store.ChannelStore, defaultModel string) *OpenHandler {
	return &OpenHandler{pipeline: p, historyStore: hs, userStore: us, channelStore: cs, defaultModel: defaultModel}
}

type OpenGenerateRequest struct {
	Input         string `json:"input" binding:"required"`
	Model         string `json:"model"`
	Mode          string `json:"mode"`
	WebhookURL    string `json:"webhook_url"`
	WebhookSecret string `json:"webhook_secret"`
}

func (h *OpenHandler) Generate(c *gin.Context) {
	var req OpenGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model == "" {
		req.Model = h.defaultModel
	}
	if req.Mode == "" {
		req.Mode = "async"
	}

	cm, err := h.channelStore.GetModelByCode(req.Model)
	if err != nil || !cm.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model not available: " + req.Model})
		return
	}

	userID := c.GetInt64("user_id")
	credits := cm.CreditsPerCall

	if err := h.userStore.DeductCredit(userID, credits); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient credits"})
		return
	}

	pipeReq := service.PipelineRequest{
		Input:      req.Input,
		UserID:     userID,
		Model:      req.Model,
		Source:     "api",
		WebhookURL: req.WebhookURL,
	}

	historyID, err := h.pipeline.ExecuteAsync(pipeReq)
	if err != nil {
		h.userStore.AddCredits(userID, credits)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Mode == "sync" {
		result := h.waitForResult(c.Request.Context(), historyID)
		if result == nil {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timeout"})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"task_id": historyID})
}

func (h *OpenHandler) waitForResult(ctx context.Context, historyID int64) *gin.H {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	ch := h.pipeline.EventBus().Subscribe(historyID)
	defer h.pipeline.EventBus().Unsubscribe(historyID, ch)

	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-ch:
			if !ok {
				return nil
			}
			if evt.Name == "done" {
				history, err := h.historyStore.GetByID(historyID)
				if err != nil {
					return nil
				}
				return &gin.H{
					"task_id":          history.ID,
					"status":           history.Status,
					"reviewer_output":  json.RawMessage(history.ReviewerOutput),
					"generator_output": json.RawMessage(history.GeneratorOutput),
					"duration_ms":      history.DurationMs,
				}
			}
			if evt.Status == "failed" {
				return &gin.H{"task_id": historyID, "status": "failed", "error": evt.Error}
			}
		}
	}
}

func (h *OpenHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	history, err := h.historyStore.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	userID := c.GetInt64("user_id")
	if history.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":          history.ID,
		"status":           history.Status,
		"current_step":     history.CurrentStep,
		"error":            history.ErrorMsg,
		"reviewer_output":  history.ReviewerOutput,
		"generator_output": history.GeneratorOutput,
		"duration_ms":      history.DurationMs,
		"created_at":       history.CreatedAt,
	})
}

func (h *OpenHandler) CancelTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.historyStore.Cancel(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found or not running"})
		return
	}

	h.pipeline.Cancel(id)
	_ = h.userStore.AddCredits(userID, 1)
	c.JSON(http.StatusOK, gin.H{"message": "cancelled"})
}
