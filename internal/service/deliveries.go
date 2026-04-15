package service

import (
	"context"
	"strings"

	"github.com/julianlinaress/service_hub_notifier/internal/domain"
)

type Transport interface {
	Deliver(ctx context.Context, req domain.DeliveryRequest) domain.DeliveryResponse
}

type DeliveryService struct {
	telegram Transport
	slack    Transport
}

func NewDeliveryService(telegram Transport, slack Transport) *DeliveryService {
	return &DeliveryService{telegram: telegram, slack: slack}
}

func (s *DeliveryService) Deliver(ctx context.Context, req domain.DeliveryRequest) domain.DeliveryResponse {
	if strings.TrimSpace(req.DeliveryAttemptKey) == "" {
		return domain.FailureResponse(false, "invalid_request", "missing delivery_attempt_key", "", nil)
	}

	if strings.TrimSpace(req.Provider) == "" {
		return domain.FailureResponse(false, "invalid_request", "missing provider", "", nil)
	}

	switch req.Provider {
	case "telegram":
		return s.telegram.Deliver(ctx, req)
	case "slack":
		return s.slack.Deliver(ctx, req)
	default:
		return domain.FailureResponse(false, "unsupported_provider", "unsupported provider", "", nil)
	}
}
