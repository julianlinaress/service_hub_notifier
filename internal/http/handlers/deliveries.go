package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/julianlinaress/service_hub_notifier/internal/adapters/logger"
	"github.com/julianlinaress/service_hub_notifier/internal/domain"
)

const maxDeliveryRequestBodyBytes = 1 << 20

type DeliveriesHandler struct {
	service deliveryUseCase
	logger  *logger.Logger
}

type deliveryUseCase interface {
	Deliver(ctx context.Context, req domain.DeliveryRequest) domain.DeliveryResponse
}

func NewDeliveriesHandler(service deliveryUseCase, logger *logger.Logger) *DeliveriesHandler {
	return &DeliveriesHandler{service: service, logger: logger}
}

func (h *DeliveriesHandler) HandleCreateDelivery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxDeliveryRequestBodyBytes)

	var req domain.DeliveryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			h.logger.Error("delivery_request_too_large", map[string]any{"error": err.Error()})
			h.writeJSON(w, http.StatusRequestEntityTooLarge, domain.FailureResponse(false, domain.ErrPayloadTooLarge, "request body too large", "413", nil))
			return
		}

		h.logger.Error("invalid_delivery_request", map[string]any{"error": err.Error()})
		h.writeJSON(w, http.StatusBadRequest, domain.FailureResponse(false, domain.ErrInvalidJSON, "invalid request body", "400", nil))
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.logger.Error("invalid_delivery_request", map[string]any{"error": "request body must contain a single JSON object"})
		h.writeJSON(w, http.StatusBadRequest, domain.FailureResponse(false, domain.ErrInvalidJSON, "invalid request body", "400", nil))
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

	if response.Status == domain.StatusDelivered {
		h.logger.Info("delivery_completed", logFields)
		h.writeJSON(w, http.StatusOK, response)
		return
	}

	h.logger.Error("delivery_failed", logFields)

	if response.ErrorCode == domain.ErrInvalidRequest || response.ErrorCode == domain.ErrUnsupportedProvider || response.ErrorCode == domain.ErrInvalidDestination || response.ErrorCode == domain.ErrInvalidJSON {
		h.writeJSON(w, http.StatusBadRequest, response)
		return
	}

	if response.ErrorCode == domain.ErrPayloadTooLarge {
		h.writeJSON(w, http.StatusRequestEntityTooLarge, response)
		return
	}

	if response.Retryable {
		h.writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	h.writeJSON(w, http.StatusUnprocessableEntity, response)
}

func (h *DeliveriesHandler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("response_write_failed", map[string]any{"error": err.Error()})
	}
}
