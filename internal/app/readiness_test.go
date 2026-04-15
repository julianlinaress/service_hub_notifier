package app

import (
	"testing"
	"time"

	"github.com/julianlinaress/service_hub_notifier/internal/config"
)

func TestReadinessReadyWhenTokenPresent(t *testing.T) {
	t.Parallel()

	readiness := NewReadiness(config.Config{
		Port:                 "8081",
		DeliveryTimeout:      2 * time.Second,
		ShutdownTimeout:      3 * time.Second,
		TelegramAPIBaseURL:   "https://api.telegram.org",
		InternalServiceToken: "internal-token",
	})

	ready, checks := readiness.Ready()
	if !ready {
		t.Fatalf("ready = %t, want true", ready)
	}

	if checks["internal_service_token_present"] != true {
		t.Fatalf("token check = %v, want true", checks["internal_service_token_present"])
	}
}

func TestReadinessNotReadyWhenTokenMissing(t *testing.T) {
	t.Parallel()

	readiness := NewReadiness(config.Config{
		Port:               "8081",
		DeliveryTimeout:    2 * time.Second,
		ShutdownTimeout:    3 * time.Second,
		TelegramAPIBaseURL: "https://api.telegram.org",
	})

	ready, checks := readiness.Ready()
	if ready {
		t.Fatalf("ready = %t, want false", ready)
	}

	if checks["internal_service_token_present"] != false {
		t.Fatalf("token check = %v, want false", checks["internal_service_token_present"])
	}
}
