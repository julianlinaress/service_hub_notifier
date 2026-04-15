package router

import (
	"net/http"
)

type deliveriesHandler interface {
	HandleCreateDelivery(w http.ResponseWriter, r *http.Request)
}

// New builds the HTTP router for service endpoints.
func New(deliveriesHandler deliveriesHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			return
		}
	})

	mux.HandleFunc("/api/v1/deliveries", deliveriesHandler.HandleCreateDelivery)

	return mux
}
