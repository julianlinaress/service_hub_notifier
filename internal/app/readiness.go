package app

import (
	"strings"

	"github.com/julianlinaress/service_hub_notifier/internal/config"
)

type Readiness struct {
	cfg config.Config
}

func NewReadiness(cfg config.Config) *Readiness {
	return &Readiness{cfg: cfg}
}

func (r *Readiness) Ready() (bool, map[string]any) {
	checks := map[string]any{
		"config_loaded":                    true,
		"port_configured":                  strings.TrimSpace(r.cfg.Port) != "",
		"delivery_timeout_valid":           r.cfg.DeliveryTimeout > 0,
		"shutdown_timeout_valid":           r.cfg.ShutdownTimeout > 0,
		"telegram_api_base_url_configured": strings.TrimSpace(r.cfg.TelegramAPIBaseURL) != "",
		"internal_service_token_present":   strings.TrimSpace(r.cfg.InternalServiceToken) != "",
		"provider_clients_initialized":     true,
	}

	ready := true
	for _, value := range checks {
		asBool, ok := value.(bool)
		if !ok || !asBool {
			ready = false
			break
		}
	}

	return ready, checks
}
