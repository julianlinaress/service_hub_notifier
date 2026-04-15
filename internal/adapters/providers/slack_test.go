package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julianlinaress/service_hub_notifier/internal/domain"
)

func TestSlackAdapterDeliver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		destination  map[string]any
		responseCode int
		responseBody string
		wantStatus   string
		wantCode     string
		wantRetry    bool
		wantRaw      string
		wantHit      bool
	}{
		{
			name:        "missing webhook url",
			destination: map[string]any{},
			wantStatus:  domain.StatusFailed,
			wantCode:    domain.ErrInvalidDestination,
		},
		{
			name:         "success json response",
			responseCode: http.StatusOK,
			responseBody: `{"ok":true}`,
			wantStatus:   domain.StatusDelivered,
			wantCode:     "",
			wantHit:      true,
		},
		{
			name:         "success text response",
			responseCode: http.StatusOK,
			responseBody: "ok",
			wantStatus:   domain.StatusDelivered,
			wantCode:     "",
			wantRaw:      "ok",
			wantHit:      true,
		},
		{
			name:         "retryable on too many requests",
			responseCode: http.StatusTooManyRequests,
			responseBody: `{"error":"rate"}`,
			wantStatus:   domain.StatusFailed,
			wantCode:     domain.ErrSlackSendFailed,
			wantRetry:    true,
			wantHit:      true,
		},
		{
			name:         "non retryable on bad request",
			responseCode: http.StatusBadRequest,
			responseBody: `{"error":"bad"}`,
			wantStatus:   domain.StatusFailed,
			wantCode:     domain.ErrSlackSendFailed,
			wantRetry:    false,
			wantHit:      true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hit := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hit = true
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
				}

				w.WriteHeader(tc.responseCode)
				if _, err := w.Write([]byte(tc.responseBody)); err != nil {
					t.Fatalf("write response: %v", err)
				}
			}))
			defer server.Close()

			destination := tc.destination
			if destination == nil {
				destination = map[string]any{"webhook_url": server.URL}
			} else if _, ok := destination["webhook_url"]; !ok && tc.wantHit {
				destination["webhook_url"] = server.URL
			}

			adapter := NewSlackAdapter(server.Client())
			response := adapter.Deliver(context.Background(), domain.DeliveryRequest{
				Destination:  destination,
				Notification: baseNotification(),
			})

			if response.Status != tc.wantStatus {
				t.Fatalf("status = %q, want %q", response.Status, tc.wantStatus)
			}

			if response.ErrorCode != tc.wantCode {
				t.Fatalf("error_code = %q, want %q", response.ErrorCode, tc.wantCode)
			}

			if response.Retryable != tc.wantRetry {
				t.Fatalf("retryable = %t, want %t", response.Retryable, tc.wantRetry)
			}

			if tc.wantRaw != "" && response.ProviderResponse["raw"] != tc.wantRaw {
				t.Fatalf("provider_response.raw = %v, want %q", response.ProviderResponse["raw"], tc.wantRaw)
			}

			if hit != tc.wantHit {
				t.Fatalf("server hit = %t, want %t", hit, tc.wantHit)
			}
		})
	}
}

func TestParseSlackBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want map[string]any
	}{
		{
			name: "json payload",
			body: `{"ok":true,"channel":"alerts"}`,
			want: map[string]any{
				"ok":      true,
				"channel": "alerts",
			},
		},
		{
			name: "text payload",
			body: "ok",
			want: map[string]any{
				"raw": "ok",
			},
		},
		{
			name: "empty payload",
			body: "\n\t",
			want: map[string]any{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := parseSlackBody(strings.NewReader(tc.body))
			if len(got) != len(tc.want) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tc.want))
			}

			for key, wantVal := range tc.want {
				if got[key] != wantVal {
					t.Fatalf("got[%q] = %#v, want %#v", key, got[key], wantVal)
				}
			}
		})
	}
}

func TestFormatSlackMessage(t *testing.T) {
	t.Parallel()

	message := formatSlackMessage(domain.DeliveryRequest{Notification: baseNotification()})
	if message["text"] == "" {
		t.Fatalf("expected non-empty text")
	}

	attachments, ok := message["attachments"].([]map[string]any)
	if !ok || len(attachments) == 0 {
		t.Fatalf("expected attachments in slack payload")
	}

	fields, ok := attachments[0]["fields"].([]map[string]any)
	if !ok || len(fields) == 0 {
		t.Fatalf("expected attachment fields")
	}

	if fields[0]["title"] != "Check" {
		t.Fatalf("first field title = %v, want %q", fields[0]["title"], "Check")
	}
}

func TestParseSlackBodyJSONValues(t *testing.T) {
	t.Parallel()

	parsed := parseSlackBody(strings.NewReader(`{"ok":true,"attempt":3}`))
	if parsed["ok"] != true {
		t.Fatalf("ok = %v, want true", parsed["ok"])
	}

	attempt, ok := parsed["attempt"].(float64)
	if !ok {
		t.Fatalf("attempt type = %T, want float64", parsed["attempt"])
	}

	if attempt != 3 {
		t.Fatalf("attempt = %v, want %d", attempt, 3)
	}
}

func baseNotification() domain.NotificationInput {
	return domain.NotificationInput{
		CheckType:    "health",
		Severity:     "alert",
		Message:      "health check failed",
		DeploymentID: 2,
		Metadata: map[string]any{
			"host": "example.com",
			"env":  "prod",
		},
	}
}
