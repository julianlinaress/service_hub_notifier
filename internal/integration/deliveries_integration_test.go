package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julianlinaress/service_hub_notifier/internal/adapters/httpclient"
	"github.com/julianlinaress/service_hub_notifier/internal/adapters/logger"
	"github.com/julianlinaress/service_hub_notifier/internal/adapters/providers"
	"github.com/julianlinaress/service_hub_notifier/internal/domain"
	"github.com/julianlinaress/service_hub_notifier/internal/http/handlers"
	"github.com/julianlinaress/service_hub_notifier/internal/http/router"
	"github.com/julianlinaress/service_hub_notifier/internal/service"
)

func TestPOSTDeliveriesTelegramSuccess(t *testing.T) {
	t.Parallel()

	telegramMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":321}}`))
	}))
	defer telegramMock.Close()

	notifier := newNotifierServer(telegramMock.URL, 2*time.Second)
	defer notifier.Close()

	statusCode, body := postDelivery(t, notifier.URL+"/api/v1/deliveries", telegramRequest())
	if statusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", statusCode, http.StatusOK)
	}

	if body.Status != domain.StatusDelivered {
		t.Fatalf("response status = %q, want %q", body.Status, domain.StatusDelivered)
	}

	if body.ProviderMessageID != "321" {
		t.Fatalf("provider_message_id = %q, want %q", body.ProviderMessageID, "321")
	}
}

func TestPOSTDeliveriesTimeoutReturnsRetryableError(t *testing.T) {
	t.Parallel()

	telegramMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer telegramMock.Close()

	notifier := newNotifierServer(telegramMock.URL, 50*time.Millisecond)
	defer notifier.Close()

	statusCode, body := postDelivery(t, notifier.URL+"/api/v1/deliveries", telegramRequest())
	if statusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", statusCode, http.StatusServiceUnavailable)
	}

	if body.ErrorCode != domain.ErrProviderRequest {
		t.Fatalf("error_code = %q, want %q", body.ErrorCode, domain.ErrProviderRequest)
	}

	if !body.Retryable {
		t.Fatalf("retryable = %t, want true", body.Retryable)
	}
}

func TestPOSTDeliveriesRetryableOnProvider5xx(t *testing.T) {
	t.Parallel()

	telegramMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"ok":false}`))
	}))
	defer telegramMock.Close()

	notifier := newNotifierServer(telegramMock.URL, 2*time.Second)
	defer notifier.Close()

	statusCode, body := postDelivery(t, notifier.URL+"/api/v1/deliveries", telegramRequest())
	if statusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", statusCode, http.StatusServiceUnavailable)
	}

	if body.ErrorCode != domain.ErrTelegramSendFailed {
		t.Fatalf("error_code = %q, want %q", body.ErrorCode, domain.ErrTelegramSendFailed)
	}

	if !body.Retryable {
		t.Fatalf("retryable = %t, want true", body.Retryable)
	}
}

func TestPOSTDeliveriesSlackWithMockEndpoint(t *testing.T) {
	t.Parallel()

	slackMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer slackMock.Close()

	telegramMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer telegramMock.Close()

	notifier := newNotifierServer(telegramMock.URL, 2*time.Second)
	defer notifier.Close()

	statusCode, body := postDelivery(t, notifier.URL+"/api/v1/deliveries", slackRequest(slackMock.URL))
	if statusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", statusCode, http.StatusOK)
	}

	if body.Status != domain.StatusDelivered {
		t.Fatalf("response status = %q, want %q", body.Status, domain.StatusDelivered)
	}

	if body.ProviderResponse["raw"] != "ok" {
		t.Fatalf("provider_response.raw = %v, want %q", body.ProviderResponse["raw"], "ok")
	}
}

func TestPOSTDeliveriesRequiresInternalToken(t *testing.T) {
	t.Parallel()

	telegramMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer telegramMock.Close()

	notifier := newNotifierServerWithToken(telegramMock.URL, 2*time.Second, "integration-secret")
	defer notifier.Close()

	statusCode, body := postDeliveryWithHeaders(t, notifier.URL+"/api/v1/deliveries", telegramRequest(), map[string]string{})
	if statusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", statusCode, http.StatusUnauthorized)
	}

	if body.ErrorCode != domain.ErrUnauthorized {
		t.Fatalf("error_code = %q, want %q", body.ErrorCode, domain.ErrUnauthorized)
	}

	statusCode, body = postDeliveryWithHeaders(t, notifier.URL+"/api/v1/deliveries", telegramRequest(), map[string]string{
		"Authorization": "Bearer integration-secret",
		"X-Request-Id":  "integration-req",
		"X-Attempt-Id":  "integration-attempt",
	})
	if statusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", statusCode, http.StatusOK)
	}

	if body.Status != domain.StatusDelivered {
		t.Fatalf("response status = %q, want %q", body.Status, domain.StatusDelivered)
	}

	metricsResp, err := http.Get(notifier.URL + "/metrics")
	if err != nil {
		t.Fatalf("get /metrics: %v", err)
	}
	defer metricsResp.Body.Close()

	if metricsResp.StatusCode != http.StatusOK {
		t.Fatalf("metrics status = %d, want %d", metricsResp.StatusCode, http.StatusOK)
	}
}

func newNotifierServer(telegramBaseURL string, timeout time.Duration) *httptest.Server {
	return newNotifierServerWithToken(telegramBaseURL, timeout, "")
}

func newNotifierServerWithToken(telegramBaseURL string, timeout time.Duration, token string) *httptest.Server {
	httpClient := httpclient.New(timeout)
	telegram := providers.NewTelegramAdapter(httpClient, providers.WithTelegramAPIBaseURL(telegramBaseURL))
	slack := providers.NewSlackAdapter(httpClient)
	deliveriesHandler := handlers.NewDeliveriesHandler(service.NewDeliveryService(telegram, slack), logger.New(), token)

	return httptest.NewServer(router.New(deliveriesHandler, readinessStub{ready: true}))
}

func postDelivery(t *testing.T, url string, payload map[string]any) (int, domain.DeliveryResponse) {
	return postDeliveryWithHeaders(t, url, payload, map[string]string{})
}

func postDeliveryWithHeaders(t *testing.T, url string, payload map[string]any, headers map[string]string) (int, domain.DeliveryResponse) {
	t.Helper()

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(rawPayload))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer resp.Body.Close()

	var parsed domain.DeliveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return resp.StatusCode, parsed
}

type readinessStub struct {
	ready bool
}

func (r readinessStub) Ready() (bool, map[string]any) {
	return r.ready, map[string]any{"stub": r.ready}
}

func telegramRequest() map[string]any {
	return map[string]any{
		"delivery_attempt_key": "integration-event:telegram",
		"provider":             "telegram",
		"destination": map[string]any{
			"token":    "token-123",
			"chat_ref": "@alerts",
		},
		"notification": map[string]any{
			"event_name":    "health.alert",
			"check_type":    "health",
			"severity":      "alert",
			"message":       "health check failed",
			"service_id":    1,
			"deployment_id": 2,
			"metadata": map[string]any{
				"host": "example.internal",
				"env":  "test",
			},
		},
		"event": map[string]any{
			"id":   "integration-event-id",
			"name": "health.alert",
			"tags": map[string]any{"source": "integration"},
		},
	}
}

func slackRequest(webhookURL string) map[string]any {
	return map[string]any{
		"delivery_attempt_key": "integration-event:slack",
		"provider":             "slack",
		"destination": map[string]any{
			"webhook_url": webhookURL,
		},
		"notification": map[string]any{
			"event_name":    "health.alert",
			"check_type":    "health",
			"severity":      "alert",
			"message":       "health check failed",
			"service_id":    1,
			"deployment_id": 2,
			"metadata": map[string]any{
				"host": "example.internal",
				"env":  "test",
			},
		},
		"event": map[string]any{
			"id":   "integration-event-id",
			"name": "health.alert",
			"tags": map[string]any{"source": "integration"},
		},
	}
}
