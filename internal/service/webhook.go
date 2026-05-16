package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"meta-prompt/internal/model"
	"meta-prompt/internal/store"
)

type WebhookService struct {
	historyStore *store.HistoryStore
}

func NewWebhookService(hs *store.HistoryStore) *WebhookService {
	return &WebhookService{historyStore: hs}
}

func (s *WebhookService) Notify(history *model.History) {
	if history.WebhookURL == "" {
		return
	}

	payload, _ := json.Marshal(map[string]any{
		"task_id":          history.ID,
		"status":           history.Status,
		"reviewer_output":  history.ReviewerOutput,
		"generator_output": history.GeneratorOutput,
		"duration_ms":      history.DurationMs,
		"error":            history.ErrorMsg,
	})

	req, err := http.NewRequest("POST", history.WebhookURL, bytes.NewReader(payload))
	if err != nil {
		slog.Error("webhook request build failed", "history_id", history.ID, "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	if history.WebhookSecret != "" {
		mac := hmac.New(sha256.New, []byte(history.WebhookSecret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("webhook delivery failed", "history_id", history.ID, "err", err)
		return
	}
	resp.Body.Close()
	slog.Info("webhook delivered", "history_id", history.ID, "status", resp.StatusCode)
}
