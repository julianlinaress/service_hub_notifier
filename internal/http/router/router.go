package router

import (
	"net/http"

	"github.com/julianlinaress/service_hub_notifier/internal/http/handlers"
)

func New(deliveriesHandler *handlers.DeliveriesHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/api/v1/deliveries", deliveriesHandler.HandleCreateDelivery)

	return mux
}
