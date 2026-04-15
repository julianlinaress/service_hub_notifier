package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julianlinaress/service_hub_notifier/internal/adapters/logger"
	"github.com/julianlinaress/service_hub_notifier/internal/domain"
	"github.com/julianlinaress/service_hub_notifier/internal/service"
)

type transportStub struct {
	response domain.DeliveryResponse
	called   bool
}

func (s *transportStub) Deliver(_ context.Context, _ domain.DeliveryRequest) domain.DeliveryResponse {
	s.called = true
	return s.response
}

func TestHandleCreateDelivery(t *testing.T) {
	t.Parallel()

	makeHandler := func(telegramResp domain.DeliveryResponse) (*DeliveriesHandler, *transportStub) {
		telegram := &transportStub{response: telegramResp}
		slack := &transportStub{response: domain.SuccessResponse("", "200", map[string]any{"ok": true})}
		deliveryService := service.NewDeliveryService(telegram, slack)

		return NewDeliveriesHandler(deliveryService, logger.New()), telegram
	}

	validRequest := map[string]any{
		"delivery_attempt_key": "event-id:telegram",
		"provider":             "telegram",
		"destination": map[string]any{
			"token":    "token",
			"chat_ref": "@alerts",
		},
		"notification": map[string]any{
			"event_name":    "health.alert",
			"check_type":    "health",
			"severity":      "alert",
			"message":       "failed",
			"service_id":    1,
			"deployment_id": 2,
			"metadata": map[string]any{
				"host": "example.com",
			},
		},
		"event": map[string]any{
			"id":   "event-id",
			"name": "health.alert",
			"tags": map[string]any{"source": "automatic"},
		},
	}

	tests := []struct {
		name           string
		body           string
		serviceResp    domain.DeliveryResponse
		expectedStatus int
		expectedCode   string
		expectCalled   bool
	}{
		{
			name: "success request",
			body: mustJSON(t, validRequest),
			serviceResp: domain.SuccessResponse(
				"123",
				"200",
				map[string]any{"ok": true},
			),
			expectedStatus: http.StatusOK,
			expectedCode:   "",
			expectCalled:   true,
		},
		{
			name: "unknown field rejected",
			body: mustJSON(t, map[string]any{
				"delivery_attempt_key": "event-id:telegram",
				"provider":             "telegram",
				"destination":          map[string]any{"token": "token", "chat_ref": "@alerts"},
				"notification":         validRequest["notification"],
				"event":                validRequest["event"],
				"unknown":              "field",
			}),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_json",
			expectCalled:   false,
		},
		{
			name:           "multiple json values rejected",
			body:           mustJSON(t, validRequest) + mustJSON(t, validRequest),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_json",
			expectCalled:   false,
		},
		{
			name:           "payload too large",
			body:           mustLargePayload(t),
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectedCode:   "payload_too_large",
			expectCalled:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler, telegram := makeHandler(tc.serviceResp)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/deliveries", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			handler.HandleCreateDelivery(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.expectedStatus)
			}

			if telegram.called != tc.expectCalled {
				t.Fatalf("telegram called = %t, want %t", telegram.called, tc.expectCalled)
			}

			if tc.expectedCode != "" {
				var response domain.DeliveryResponse
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("decode response: %v", err)
				}

				if response.ErrorCode != tc.expectedCode {
					t.Fatalf("error_code = %q, want %q", response.ErrorCode, tc.expectedCode)
				}
			}
		})
	}
}

func mustJSON(t *testing.T, value map[string]any) string {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}

	return string(raw)
}

func mustLargePayload(t *testing.T) string {
	t.Helper()

	payload := map[string]any{
		"delivery_attempt_key": "event-id:telegram",
		"provider":             "telegram",
		"destination": map[string]any{
			"token":    "token",
			"chat_ref": "@alerts",
		},
		"notification": map[string]any{
			"event_name":    "health.alert",
			"check_type":    "health",
			"severity":      "alert",
			"message":       strings.Repeat("x", maxDeliveryRequestBodyBytes),
			"service_id":    1,
			"deployment_id": 2,
			"metadata":      map[string]any{},
		},
		"event": map[string]any{
			"id":   "event-id",
			"name": "health.alert",
			"tags": map[string]any{},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if len(raw) <= maxDeliveryRequestBodyBytes {
		t.Fatalf("payload length = %d, want > %d", len(raw), maxDeliveryRequestBodyBytes)
	}

	return string(raw)
}
