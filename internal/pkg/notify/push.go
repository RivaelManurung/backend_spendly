package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type NotificationType string

const (
	TypeDailyInsight  NotificationType = "daily_insight"
	TypeBudgetAlert   NotificationType = "budget_alert"
	TypeMonthlyReport NotificationType = "monthly_report"
	TypeAnomalyAlert  NotificationType = "anomaly_alert"
)

type Notification struct {
	UserID  string           `json:"user_id"`
	Type    NotificationType `json:"type"`
	Title   string           `json:"title"`
	Body    string           `json:"body"`
	Data    map[string]any   `json:"data,omitempty"`
}

type Event struct {
	Name      string         `json:"name"`
	Payload   map[string]any `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
}

type Emitter struct {
	webhookURL string
	httpClient *http.Client
	fcmKey     string
}

func New() *Emitter {
	return &Emitter{
		webhookURL: os.Getenv("INTERNAL_WEBHOOK_URL"),
		fcmKey:     os.Getenv("FCM_SERVER_KEY"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Emit fires an internal event to the webhook bus (triggers other agent workflows).
func (e *Emitter) Emit(ctx context.Context, eventName string, payload map[string]any) error {
	event := Event{
		Name:      eventName,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	body, _ := json.Marshal(event)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, e.webhookURL+"/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", os.Getenv("INTERNAL_API_KEY"))

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Emit %s: %w", eventName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Emit %s: server returned %d", eventName, resp.StatusCode)
	}

	return nil
}

// Push sends a push notification via FCM (Firebase Cloud Messaging).
func (e *Emitter) Push(ctx context.Context, n *Notification) error {
	// Fetch device tokens for user from DB (simplified here)
	tokens, err := e.getDeviceTokens(ctx, n.UserID)
	if err != nil || len(tokens) == 0 {
		return nil // no tokens = user has no devices registered, not an error
	}

	payload := map[string]any{
		"registration_ids": tokens,
		"notification": map[string]string{
			"title": n.Title,
			"body":  n.Body,
		},
		"data": n.Data,
		"android": map[string]any{
			"priority": "high",
			"notification": map[string]string{
				"channel_id": string(n.Type),
			},
		},
		"apns": map[string]any{
			"headers": map[string]string{
				"apns-priority": "10",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://fcm.googleapis.com/fcm/send", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+e.fcmKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Push FCM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Push FCM: status %d", resp.StatusCode)
	}

	return nil
}

// PushBudgetAlert is a convenience wrapper for budget threshold alerts.
func (e *Emitter) PushBudgetAlert(ctx context.Context, userID string, categoryName string, thresholdPct int, message string) error {
	return e.Push(ctx, &Notification{
		UserID: userID,
		Type:   TypeBudgetAlert,
		Title:  fmt.Sprintf("Budget %s: %d%% terpakai", categoryName, thresholdPct),
		Body:   message,
		Data: map[string]any{
			"type":          "budget_alert",
			"category":      categoryName,
			"threshold_pct": thresholdPct,
		},
	})
}

// PushDailyInsight sends the daily digest notification.
func (e *Emitter) PushDailyInsight(ctx context.Context, userID, summary string) error {
	return e.Push(ctx, &Notification{
		UserID: userID,
		Type:   TypeDailyInsight,
		Title:  "Ringkasan Harian",
		Body:   summary,
		Data:   map[string]any{"type": "daily_insight"},
	})
}

// getDeviceTokens fetches FCM tokens for a user.
// In production: query a device_tokens table.
func (e *Emitter) getDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	// Stub — replace with DB query
	return []string{}, nil
}
