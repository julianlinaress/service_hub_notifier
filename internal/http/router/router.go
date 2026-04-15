package router

import (
	"encoding/json"
	"net/http"

	"github.com/julianlinaress/service_hub_notifier/internal/metrics"
)

type deliveriesHandler interface {
	HandleCreateDelivery(w http.ResponseWriter, r *http.Request)
}

type readinessChecker interface {
	Ready() (bool, map[string]any)
}

// New builds the HTTP router for service endpoints.
func New(deliveriesHandler deliveriesHandler, readiness readinessChecker) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]any{"status": "ok"}); err != nil {
			return
		}
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		body := map[string]any{"status": "ready"}

		if readiness != nil {
			ready, checks := readiness.Ready()
			body["checks"] = checks
			if !ready {
				status = http.StatusServiceUnavailable
				body["status"] = "not_ready"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(body); err != nil {
			return
		}
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(metrics.PrometheusText())); err != nil {
			return
		}
	})

	mux.HandleFunc("/api/v1/deliveries", deliveriesHandler.HandleCreateDelivery)

	return mux
}
