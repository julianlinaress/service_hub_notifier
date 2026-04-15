package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/julianlinaress/service_hub_notifier/internal/domain"
	"github.com/julianlinaress/service_hub_notifier/internal/metrics"
)

const maxDeliveryRequestBodyBytes = 1 << 20

type DeliveriesHandler struct {
	service deliveryUseCase
	logger  eventLogger
	token   string
}

type noopLogger struct{}

func (noopLogger) Info(string, map[string]any) {}

func (noopLogger) Error(string, map[string]any) {}

type deliveryUseCase interface {
	Deliver(ctx context.Context, req domain.DeliveryRequest) domain.DeliveryResponse
}

type eventLogger interface {
	Info(message string, fields map[string]any)
	Error(message string, fields map[string]any)
}

// NewDeliveriesHandler builds the HTTP handler for delivery requests.
func NewDeliveriesHandler(service deliveryUseCase, logger eventLogger, internalToken string) *DeliveriesHandler {
	if logger == nil {
		logger = noopLogger{}
	}

	return &DeliveriesHandler{service: service, logger: logger, token: strings.TrimSpace(internalToken)}
}

// HandleCreateDelivery validates and processes a delivery request.
func (h *DeliveriesHandler) HandleCreateDelivery(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	requestID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
	attemptID := strings.TrimSpace(r.Header.Get("X-Attempt-Id"))

	if !h.authorized(r) {
		h.logger.Error("delivery_request_unauthorized", map[string]any{
			"request_id": requestID,
			"attempt_id": attemptID,
			"status":     http.StatusUnauthorized,
		})
		h.writeJSON(w, http.StatusUnauthorized, domain.FailureResponse(false, domain.ErrUnauthorized, "missing or invalid service token", "401", nil))
		return
	}

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
			h.logger.Error("delivery_request_too_large", map[string]any{
				"request_id": requestID,
				"attempt_id": attemptID,
				"error":      err.Error(),
				"latency_ms": elapsedMilliseconds(startedAt),
			})
			h.writeJSON(w, http.StatusRequestEntityTooLarge, domain.FailureResponse(false, domain.ErrPayloadTooLarge, "request body too large", "413", nil))
			return
		}

		h.logger.Error("invalid_delivery_request", map[string]any{
			"request_id": requestID,
			"attempt_id": attemptID,
			"error":      err.Error(),
			"latency_ms": elapsedMilliseconds(startedAt),
		})
		h.writeJSON(w, http.StatusBadRequest, domain.FailureResponse(false, domain.ErrInvalidJSON, "invalid request body", "400", nil))
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.logger.Error("invalid_delivery_request", map[string]any{
			"request_id": requestID,
			"attempt_id": attemptID,
			"error":      "request body must contain a single JSON object",
			"latency_ms": elapsedMilliseconds(startedAt),
		})
		h.writeJSON(w, http.StatusBadRequest, domain.FailureResponse(false, domain.ErrInvalidJSON, "invalid request body", "400", nil))
		return
	}

	response := h.service.Deliver(r.Context(), req)
	latencyMS := elapsedMilliseconds(startedAt)
	metrics.Record(req.Provider, response.Status, latencyMS)

	logFields := map[string]any{
		"request_id":           requestID,
		"attempt_id":           firstNonEmpty(attemptID, req.DeliveryAttemptKey),
		"provider":             req.Provider,
		"delivery_attempt_key": req.DeliveryAttemptKey,
		"status":               response.Status,
		"error_code":           response.ErrorCode,
		"retryable":            response.Retryable,
		"latency_ms":           latencyMS,
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

func (h *DeliveriesHandler) authorized(r *http.Request) bool {
	if h.token == "" {
		return true
	}

	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	const bearerPrefix = "Bearer "

	if !strings.HasPrefix(authorization, bearerPrefix) {
		return false
	}

	providedToken := strings.TrimSpace(strings.TrimPrefix(authorization, bearerPrefix))
	if providedToken == "" {
		return false
	}

	return providedToken == h.token
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func elapsedMilliseconds(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

func (h *DeliveriesHandler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("response_write_failed", map[string]any{"error": err.Error()})
	}
}
