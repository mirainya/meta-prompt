package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/service"
	"meta-prompt/internal/store"
)

type SSEHandler struct {
	eventBus     *service.EventBus
	historyStore *store.HistoryStore
}

func NewSSEHandler(eb *service.EventBus, hs *store.HistoryStore) *SSEHandler {
	return &SSEHandler{eventBus: eb, historyStore: hs}
}

func (h *SSEHandler) Stream(c *gin.Context) {
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

	// 已完成的直接返回 done 事件
	if history.Status == "done" || history.Status == "failed" || history.Status == "cancelled" {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		data, _ := json.Marshal(gin.H{"status": history.Status, "current_step": history.CurrentStep})
		fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", history.Status, data)
		c.Writer.Flush()
		return
	}

	ch := h.eventBus.Subscribe(id)
	defer h.eventBus.Unsubscribe(id, ch)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Writer.Flush()

	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt)
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", evt.Status, data)
			c.Writer.Flush()
			if evt.Name == "done" || evt.Status == "failed" {
				return
			}
		}
	}
}
