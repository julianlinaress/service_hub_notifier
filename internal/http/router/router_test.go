package router

import (
	"context"
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

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	h := handlers.NewDeliveriesHandler(&deliveryUseCaseStub{}, logger.New())
	r := New(h)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want %q", got, "application/json")
	}

	if got := rec.Body.String(); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}
