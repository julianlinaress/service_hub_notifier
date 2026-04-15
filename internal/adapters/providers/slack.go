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

type SlackAdapter struct {
	httpClient *http.Client
}

func NewSlackAdapter(httpClient *http.Client) *SlackAdapter {
	return &SlackAdapter{httpClient: httpClient}
}

func (a *SlackAdapter) Deliver(ctx context.Context, req domain.DeliveryRequest) domain.DeliveryResponse {
	webhookURL, _ := req.Destination["webhook_url"].(string)

	if strings.TrimSpace(webhookURL) == "" {
		return domain.FailureResponse(false, "invalid_destination", "missing slack webhook_url", "", nil)
	}

	payload := formatSlackMessage(req)
	body, err := json.Marshal(payload)
	if err != nil {
		return domain.FailureResponse(false, "encoding_error", err.Error(), "", nil)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return domain.FailureResponse(false, "request_build_failed", err.Error(), "", nil)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return domain.FailureResponse(true, "provider_request_failed", err.Error(), "", nil)
	}
	defer resp.Body.Close()

	providerBody := parseSlackBody(resp.Body)
	code := fmt.Sprintf("%d", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		return domain.SuccessResponse("", code, providerBody)
	}

	retryable := resp.StatusCode == 429 || resp.StatusCode >= 500
	return domain.FailureResponse(retryable, "slack_send_failed", "slack webhook returned non-success status", code, providerBody)
}

func formatSlackMessage(req domain.DeliveryRequest) map[string]any {
	emoji := severityEmoji(req.Notification.Severity)
	host := stringFromAny(req.Notification.Metadata["host"], "unknown")
	env := stringFromAny(req.Notification.Metadata["env"], "unknown")

	return map[string]any{
		"text": fmt.Sprintf("%s %s: %s", emoji, strings.ToUpper(req.Notification.Severity), req.Notification.Message),
		"attachments": []map[string]any{
			{
				"color": severityColor(req.Notification.Severity),
				"fields": []map[string]any{
					{"title": "Check", "value": req.Notification.CheckType, "short": true},
					{"title": "Deployment", "value": stringFromAny(req.Notification.DeploymentID, "unknown"), "short": true},
					{"title": "Host", "value": host, "short": true},
					{"title": "Env", "value": env, "short": true},
				},
			},
		},
	}
}

func parseSlackBody(body io.Reader) map[string]any {
	raw, err := io.ReadAll(body)
	if err != nil {
		return map[string]any{}
	}

	text := strings.TrimSpace(string(raw))
	if text == "" {
		return map[string]any{}
	}

	decoded := map[string]any{}
	if err := json.Unmarshal(raw, &decoded); err == nil {
		return decoded
	}

	return map[string]any{"raw": text}
}

func severityColor(severity string) string {
	switch severity {
	case "alert":
		return "danger"
	case "warning":
		return "warning"
	case "recovery":
		return "good"
	default:
		return "#36a64f"
	}
}
