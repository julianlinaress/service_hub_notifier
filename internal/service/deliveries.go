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
	providers map[string]Transport
}

func NewDeliveryService(telegram Transport, slack Transport) *DeliveryService {
	return &DeliveryService{
		providers: map[string]Transport{
			domain.ProviderTelegram: telegram,
			domain.ProviderSlack:    slack,
		},
	}
}

func (s *DeliveryService) Deliver(ctx context.Context, req domain.DeliveryRequest) domain.DeliveryResponse {
	if strings.TrimSpace(req.DeliveryAttemptKey) == "" {
		return domain.FailureResponse(false, domain.ErrInvalidRequest, "missing delivery_attempt_key", "", nil)
	}

	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		return domain.FailureResponse(false, domain.ErrInvalidRequest, "missing provider", "", nil)
	}

	transport, ok := s.providers[provider]
	if !ok {
		return domain.FailureResponse(false, domain.ErrUnsupportedProvider, "unsupported provider", "", nil)
	}

	req.Provider = provider

	return transport.Deliver(ctx, req)
}
