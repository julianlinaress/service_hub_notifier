package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/julianlinaress/service_hub_notifier/internal/config"
)

func TestBuildServerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		wantPort string
		wantTO   time.Duration
		wantSDTO time.Duration
		wantTG   string
	}{
		{
			name:     "defaults",
			env:      map[string]string{},
			wantPort: "8081",
			wantTO:   5 * time.Second,
			wantSDTO: 10 * time.Second,
			wantTG:   "https://api.telegram.org",
		},
		{
			name: "configured",
			env: map[string]string{
				"PORT":                  "9000",
				"DELIVERY_TIMEOUT_MS":   "3000",
				"SHUTDOWN_TIMEOUT_MS":   "4000",
				"TELEGRAM_API_BASE_URL": "https://telegram.mock",
			},
			wantPort: "9000",
			wantTO:   3 * time.Second,
			wantSDTO: 4 * time.Second,
			wantTG:   "https://telegram.mock",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.LoadFromEnv(func(key string) string {
				if value, ok := tc.env[key]; ok {
					return value
				}

				return ""
			})

			if cfg.Port != tc.wantPort {
				t.Fatalf("port = %q, want %q", cfg.Port, tc.wantPort)
			}

			if cfg.DeliveryTimeout != tc.wantTO {
				t.Fatalf("delivery timeout = %s, want %s", cfg.DeliveryTimeout, tc.wantTO)
			}

			if cfg.ShutdownTimeout != tc.wantSDTO {
				t.Fatalf("shutdown timeout = %s, want %s", cfg.ShutdownTimeout, tc.wantSDTO)
			}

			if cfg.TelegramAPIBaseURL != tc.wantTG {
				t.Fatalf("telegram api base = %q, want %q", cfg.TelegramAPIBaseURL, tc.wantTG)
			}
		})
	}
}

func TestRunReturnsOnShutdownSignal(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"PORT":                "18081",
		"DELIVERY_TIMEOUT_MS": "1000",
		"SHUTDOWN_TIMEOUT_MS": "1000",
	}

	shutdown := make(chan os.Signal, 1)
	shutdown <- syscall.SIGTERM

	if err := run(func(key string) string {
		if value, ok := env[key]; ok {
			return value
		}

		return ""
	}, shutdown); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}
