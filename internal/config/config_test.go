package config

import (
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
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
			name:     "uses defaults when env missing",
			env:      map[string]string{},
			wantPort: "8081",
			wantTO:   5 * time.Second,
			wantSDTO: 10 * time.Second,
			wantTG:   "https://api.telegram.org",
		},
		{
			name: "uses configured values",
			env: map[string]string{
				"PORT":                  "9090",
				"DELIVERY_TIMEOUT_MS":   "1500",
				"SHUTDOWN_TIMEOUT_MS":   "2500",
				"TELEGRAM_API_BASE_URL": "https://telegram.mock",
			},
			wantPort: "9090",
			wantTO:   1500 * time.Millisecond,
			wantSDTO: 2500 * time.Millisecond,
			wantTG:   "https://telegram.mock",
		},
		{
			name: "falls back on invalid numbers",
			env: map[string]string{
				"DELIVERY_TIMEOUT_MS": "0",
				"SHUTDOWN_TIMEOUT_MS": "-10",
			},
			wantPort: "8081",
			wantTO:   5 * time.Second,
			wantSDTO: 10 * time.Second,
			wantTG:   "https://api.telegram.org",
		},
		{
			name: "trims port whitespace",
			env: map[string]string{
				"PORT": " 7070 ",
			},
			wantPort: "7070",
			wantTO:   5 * time.Second,
			wantSDTO: 10 * time.Second,
			wantTG:   "https://api.telegram.org",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := LoadFromEnv(func(key string) string {
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
