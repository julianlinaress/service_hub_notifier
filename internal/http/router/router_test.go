package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julianlinaress/service_hub_notifier/internal/adapters/logger"
	"github.com/julianlinaress/service_hub_notifier/internal/domain"
	"github.com/julianlinaress/service_hub_notifier/internal/http/handlers"
)

type deliveryUseCaseStub struct{}

func (s *deliveryUseCaseStub) Deliver(_ context.Context, _ domain.DeliveryRequest) domain.DeliveryResponse {
	return domain.SuccessResponse("", "200", map[string]any{"ok": true})
}

type readinessStub struct {
	ready bool
}

func (r readinessStub) Ready() (bool, map[string]any) {
	return r.ready, map[string]any{"stub": r.ready}
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	h := handlers.NewDeliveriesHandler(&deliveryUseCaseStub{}, logger.New(), "")
	r := New(h, readinessStub{ready: true})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want %q", got, "application/json")
	}

	var payload map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("status payload = %q, want %q", payload["status"], "ok")
	}
}

func TestReadyEndpoint(t *testing.T) {
	t.Parallel()

	h := handlers.NewDeliveriesHandler(&deliveryUseCaseStub{}, logger.New(), "")
	r := New(h, readinessStub{ready: true})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["status"] != "ready" {
		t.Fatalf("status payload = %v, want %q", payload["status"], "ready")
	}
}

func TestReadyEndpointNotReady(t *testing.T) {
	t.Parallel()

	h := handlers.NewDeliveriesHandler(&deliveryUseCaseStub{}, logger.New(), "")
	r := New(h, readinessStub{ready: false})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	t.Parallel()

	h := handlers.NewDeliveriesHandler(&deliveryUseCaseStub{}, logger.New(), "")
	r := New(h, readinessStub{ready: true})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "text/plain; version=0.0.4" {
		t.Fatalf("content-type = %q, want %q", got, "text/plain; version=0.0.4")
	}
}
