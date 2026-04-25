package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/store"
)

type HistoryHandler struct {
	store *store.HistoryStore
}

func NewHistoryHandler(s *store.HistoryStore) *HistoryHandler {
	return &HistoryHandler{store: s}
}

func (h *HistoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	history, err := h.store.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// 校验归属
	userID := c.GetInt64("user_id")
	if history.UserID != 0 && history.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *HistoryHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	userID := c.GetInt64("user_id")

	list, err := h.store.ListByUser(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Running 查询用户当前正在进行的推演
func (h *HistoryHandler) Running(c *gin.Context) {
	userID := c.GetInt64("user_id")
	history, err := h.store.GetRunningByUser(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"running": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"running":      true,
		"id":           history.ID,
		"current_step": history.CurrentStep,
		"input":        history.Input,
	})
}
