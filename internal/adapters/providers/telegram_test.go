package providers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julianlinaress/service_hub_notifier/internal/domain"
)

func TestTelegramAdapterDeliver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		destination  map[string]any
		responseCode int
		responseBody string
		wantStatus   string
		wantCode     string
		wantRetry    bool
		wantMessage  string
		wantHit      bool
	}{
		{
			name:        "missing token",
			destination: map[string]any{"chat_ref": "@alerts"},
			wantStatus:  domain.StatusFailed,
			wantCode:    domain.ErrInvalidDestination,
		},
		{
			name:        "missing chat ref",
			destination: map[string]any{"token": "abc"},
			wantStatus:  domain.StatusFailed,
			wantCode:    domain.ErrInvalidDestination,
		},
		{
			name:         "success response extracts message id",
			responseCode: http.StatusOK,
			responseBody: `{"ok":true,"result":{"message_id":42}}`,
			wantStatus:   domain.StatusDelivered,
			wantCode:     "",
			wantMessage:  "42",
			wantHit:      true,
		},
		{
			name:         "retryable on 429",
			responseCode: http.StatusTooManyRequests,
			responseBody: `{"ok":false}`,
			wantStatus:   domain.StatusFailed,
			wantCode:     domain.ErrTelegramSendFailed,
			wantRetry:    true,
			wantHit:      true,
		},
		{
			name:         "non retryable on 400",
			responseCode: http.StatusBadRequest,
			responseBody: `{"ok":false}`,
			wantStatus:   domain.StatusFailed,
			wantCode:     domain.ErrTelegramSendFailed,
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
				destination = map[string]any{"token": "abc", "chat_ref": "@alerts"}
			}

			adapter := NewTelegramAdapter(server.Client(), WithTelegramAPIBaseURL(server.URL))
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

			if response.ProviderMessageID != tc.wantMessage {
				t.Fatalf("provider_message_id = %q, want %q", response.ProviderMessageID, tc.wantMessage)
			}

			if hit != tc.wantHit {
				t.Fatalf("server hit = %t, want %t", hit, tc.wantHit)
			}
		})
	}
}

func TestFormatTelegramMessage(t *testing.T) {
	t.Parallel()

	got := formatTelegramMessage(domain.DeliveryRequest{Notification: baseNotification()})
	if !strings.Contains(got, "<b>Check:</b>") {
		t.Fatalf("message missing check section: %q", got)
	}

	if !strings.Contains(got, "health check failed") {
		t.Fatalf("message missing notification text: %q", got)
	}
}

func TestSanitizeProviderError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		want    string
		wantNot string
	}{
		{
			name:    "redacts telegram token in url",
			err:     errors.New("Post \"https://api.telegram.org/bot123456:ABC/sendMessage\": dial tcp timeout"),
			want:    "Post \"https://api.telegram.org/bot<redacted>/sendMessage\": dial tcp timeout",
			wantNot: "bot123456:ABC",
		},
		{
			name:    "keeps non telegram errors",
			err:     errors.New("dial tcp 10.0.0.1:443: connect: refused"),
			want:    "dial tcp 10.0.0.1:443: connect: refused",
			wantNot: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeProviderError(tc.err)
			if got != tc.want {
				t.Fatalf("sanitizeProviderError() = %q, want %q", got, tc.want)
			}

			if tc.wantNot != "" && strings.Contains(got, tc.wantNot) {
				t.Fatalf("sanitizeProviderError() leaked token: %q", got)
			}
		})
	}
}
