package service

import (
	"context"
	"testing"

	"github.com/julianlinaress/service_hub_notifier/internal/domain"
)

type transportSpy struct {
	called bool
	req    domain.DeliveryRequest
	resp   domain.DeliveryResponse
}

func (s *transportSpy) Deliver(_ context.Context, req domain.DeliveryRequest) domain.DeliveryResponse {
	s.called = true
	s.req = req

	return s.resp
}

func TestDeliveryServiceDeliver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		request          domain.DeliveryRequest
		expectedStatus   string
		expectedCode     string
		expectTelegram   bool
		expectSlack      bool
		expectedProvider string
	}{
		{
			name: "rejects missing delivery attempt key",
			request: domain.DeliveryRequest{
				Provider: "telegram",
			},
			expectedStatus:   domain.StatusFailed,
			expectedCode:     domain.ErrInvalidRequest,
			expectTelegram:   false,
			expectSlack:      false,
			expectedProvider: "",
		},
		{
			name: "rejects missing provider",
			request: domain.DeliveryRequest{
				DeliveryAttemptKey: "event-id:provider",
			},
			expectedStatus:   domain.StatusFailed,
			expectedCode:     domain.ErrInvalidRequest,
			expectTelegram:   false,
			expectSlack:      false,
			expectedProvider: "",
		},
		{
			name: "normalizes provider name and routes telegram",
			request: domain.DeliveryRequest{
				DeliveryAttemptKey: "event-id:telegram",
				Provider:           " TeLeGrAm ",
			},
			expectedStatus:   domain.StatusDelivered,
			expectedCode:     "",
			expectTelegram:   true,
			expectSlack:      false,
			expectedProvider: domain.ProviderTelegram,
		},
		{
			name: "routes slack provider",
			request: domain.DeliveryRequest{
				DeliveryAttemptKey: "event-id:slack",
				Provider:           domain.ProviderSlack,
			},
			expectedStatus:   domain.StatusDelivered,
			expectedCode:     "",
			expectTelegram:   false,
			expectSlack:      true,
			expectedProvider: domain.ProviderSlack,
		},
		{
			name: "rejects unsupported provider",
			request: domain.DeliveryRequest{
				DeliveryAttemptKey: "event-id:email",
				Provider:           "email",
			},
			expectedStatus:   domain.StatusFailed,
			expectedCode:     domain.ErrUnsupportedProvider,
			expectTelegram:   false,
			expectSlack:      false,
			expectedProvider: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			telegram := &transportSpy{resp: domain.SuccessResponse("", "200", map[string]any{"ok": true})}
			slack := &transportSpy{resp: domain.SuccessResponse("", "200", map[string]any{"ok": true})}
			svc := NewDeliveryService(telegram, slack)

			got := svc.Deliver(context.Background(), tc.request)

			if got.Status != tc.expectedStatus {
				t.Fatalf("status = %q, want %q", got.Status, tc.expectedStatus)
			}

			if got.ErrorCode != tc.expectedCode {
				t.Fatalf("error_code = %q, want %q", got.ErrorCode, tc.expectedCode)
			}

			if telegram.called != tc.expectTelegram {
				t.Fatalf("telegram called = %t, want %t", telegram.called, tc.expectTelegram)
			}

			if slack.called != tc.expectSlack {
				t.Fatalf("slack called = %t, want %t", slack.called, tc.expectSlack)
			}

			if tc.expectTelegram && telegram.req.Provider != tc.expectedProvider {
				t.Fatalf("telegram provider = %q, want %q", telegram.req.Provider, tc.expectedProvider)
			}

			if tc.expectSlack && slack.req.Provider != tc.expectedProvider {
				t.Fatalf("slack provider = %q, want %q", slack.req.Provider, tc.expectedProvider)
			}
		})
	}
}
