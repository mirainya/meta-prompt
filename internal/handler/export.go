package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/store"
)

type ExportHandler struct {
	historyStore *store.HistoryStore
}

func NewExportHandler(hs *store.HistoryStore) *ExportHandler {
	return &ExportHandler{historyStore: hs}
}

func (h *ExportHandler) Export(c *gin.Context) {
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
	if history.UserID != 0 && history.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if history.Status != "done" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task not completed"})
		return
	}

	format := c.DefaultQuery("format", "markdown")

	switch format {
	case "json":
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=prompts_%d.json", id))
		c.Data(http.StatusOK, "application/json", history.ReviewerOutput)
	default:
		md := buildMarkdown(history.Input, history.ReviewerOutput)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=prompts_%d.md", id))
		c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(md))
	}
}

func buildMarkdown(input string, reviewerOutput json.RawMessage) string {
	var sb strings.Builder
	sb.WriteString("# 提示词生成结果\n\n")
	sb.WriteString(fmt.Sprintf("## 原始需求\n\n%s\n\n", input))
	sb.WriteString("## 生成的提示词\n\n")

	var result struct {
		Prompts []struct {
			Name           string `json:"name"`
			PromptText     string `json:"prompt_text"`
			UserInstruction string `json:"user_instruction"`
		} `json:"prompts"`
	}
	if err := json.Unmarshal(reviewerOutput, &result); err != nil {
		sb.WriteString("```json\n")
		var pretty json.RawMessage
		if json.Unmarshal(reviewerOutput, &pretty) == nil {
			formatted, _ := json.MarshalIndent(pretty, "", "  ")
			sb.Write(formatted)
		}
		sb.WriteString("\n```\n")
		return sb.String()
	}

	for i, p := range result.Prompts {
		sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, p.Name))
		if p.UserInstruction != "" {
			sb.WriteString(fmt.Sprintf("**使用说明：** %s\n\n", p.UserInstruction))
		}
		sb.WriteString("```\n")
		sb.WriteString(p.PromptText)
		sb.WriteString("\n```\n\n")
	}

	return sb.String()
}
