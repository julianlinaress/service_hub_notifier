package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julianlinaress/service_hub_notifier/internal/domain"
)

type TelegramAdapter struct {
	httpClient *http.Client
}

func NewTelegramAdapter(httpClient *http.Client) *TelegramAdapter {
	return &TelegramAdapter{httpClient: httpClient}
}

func (a *TelegramAdapter) Deliver(ctx context.Context, req domain.DeliveryRequest) domain.DeliveryResponse {
	token, _ := req.Destination["token"].(string)
	chatRef, _ := req.Destination["chat_ref"].(string)
	parseMode, _ := req.Destination["parse_mode"].(string)
	threadID := req.Destination["thread_id"]

	if strings.TrimSpace(token) == "" {
		return domain.FailureResponse(false, "invalid_destination", "missing telegram token", "", nil)
	}

	if strings.TrimSpace(chatRef) == "" {
		return domain.FailureResponse(false, "invalid_destination", "missing telegram chat_ref", "", nil)
	}

	if strings.TrimSpace(parseMode) == "" {
		parseMode = "HTML"
	}

	payload := map[string]any{
		"chat_id":    chatRef,
		"text":       formatTelegramMessage(req),
		"parse_mode": parseMode,
	}

	if threadID != nil {
		payload["message_thread_id"] = threadID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return domain.FailureResponse(false, "encoding_error", err.Error(), "", nil)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return domain.FailureResponse(false, "request_build_failed", err.Error(), "", nil)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return domain.FailureResponse(true, "provider_request_failed", err.Error(), "", nil)
	}
	defer resp.Body.Close()

	providerBody := parseBody(resp.Body)
	code := fmt.Sprintf("%d", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		messageID := extractTelegramMessageID(providerBody)
		return domain.SuccessResponse(messageID, code, providerBody)
	}

	retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
	return domain.FailureResponse(retryable, "telegram_send_failed", "telegram API returned non-success status", code, providerBody)
}

func formatTelegramMessage(req domain.DeliveryRequest) string {
	emoji := severityEmoji(req.Notification.Severity)
	host := stringFromAny(req.Notification.Metadata["host"], "unknown")
	env := stringFromAny(req.Notification.Metadata["env"], "unknown")

	return fmt.Sprintf(
		"%s <b>%s</b>\n\n<b>Check:</b> %s\n<b>Deployment:</b> %s\n<b>Host:</b> %s\n<b>Env:</b> %s\n\n%s",
		emoji,
		strings.ToUpper(req.Notification.Severity),
		req.Notification.CheckType,
		stringFromAny(req.Notification.DeploymentID, "unknown"),
		host,
		env,
		req.Notification.Message,
	)
}

func extractTelegramMessageID(body map[string]any) string {
	result, ok := body["result"].(map[string]any)
	if !ok {
		return ""
	}

	return stringFromAny(result["message_id"], "")
}

func parseBody(body io.Reader) map[string]any {
	decoded := map[string]any{}

	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return map[string]any{"raw": ""}
	}

	return decoded
}

func severityEmoji(severity string) string {
	switch severity {
	case "alert":
		return "🚨"
	case "warning":
		return "⚠️"
	case "recovery":
		return "✅"
	case "info":
		return "ℹ️"
	default:
		return "📢"
	}
}

func stringFromAny(value any, fallback string) string {
	if value == nil {
		return fallback
	}

	if asString, ok := value.(string); ok {
		trimmed := strings.TrimSpace(asString)
		if trimmed == "" {
			return fallback
		}

		return trimmed
	}

	return fmt.Sprintf("%v", value)
}
