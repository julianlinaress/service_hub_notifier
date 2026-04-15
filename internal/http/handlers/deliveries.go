package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/julianlinaress/service_hub_notifier/internal/adapters/logger"
	"github.com/julianlinaress/service_hub_notifier/internal/domain"
	"github.com/julianlinaress/service_hub_notifier/internal/service"
)

type DeliveriesHandler struct {
	service *service.DeliveryService
	logger  *logger.Logger
}

func NewDeliveriesHandler(service *service.DeliveryService, logger *logger.Logger) *DeliveriesHandler {
	return &DeliveriesHandler{service: service, logger: logger}
}

func (h *DeliveriesHandler) HandleCreateDelivery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req domain.DeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid_delivery_request", map[string]any{"error": err.Error()})
		writeJSON(w, http.StatusBadRequest, domain.FailureResponse(false, "invalid_json", "invalid request body", "400", nil))
		return
	}

	response := h.service.Deliver(r.Context(), req)

	logFields := map[string]any{
		"provider":             req.Provider,
		"delivery_attempt_key": req.DeliveryAttemptKey,
		"status":               response.Status,
		"error_code":           response.ErrorCode,
		"retryable":            response.Retryable,
	}

	if response.Status == "delivered" {
		h.logger.Info("delivery_completed", logFields)
		writeJSON(w, http.StatusOK, response)
		return
	}

	h.logger.Error("delivery_failed", logFields)

	if response.ErrorCode == "invalid_request" || response.ErrorCode == "unsupported_provider" || response.ErrorCode == "invalid_destination" || response.ErrorCode == "invalid_json" {
		writeJSON(w, http.StatusBadRequest, response)
		return
	}

	if response.Retryable {
		writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	writeJSON(w, http.StatusUnprocessableEntity, response)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
