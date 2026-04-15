package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/julianlinaress/service_hub_notifier/internal/adapters/httpclient"
	"github.com/julianlinaress/service_hub_notifier/internal/adapters/logger"
	"github.com/julianlinaress/service_hub_notifier/internal/adapters/providers"
	"github.com/julianlinaress/service_hub_notifier/internal/http/handlers"
	"github.com/julianlinaress/service_hub_notifier/internal/http/router"
	"github.com/julianlinaress/service_hub_notifier/internal/service"
)

func main() {
	port := envOrDefault("PORT", "8081")
	timeout := timeoutFromEnv("DELIVERY_TIMEOUT_MS", 5000)
	log := logger.New()

	httpClient := httpclient.New(timeout)
	telegram := providers.NewTelegramAdapter(httpClient)
	slack := providers.NewSlackAdapter(httpClient)

	deliveryService := service.NewDeliveryService(telegram, slack)
	deliveriesHandler := handlers.NewDeliveriesHandler(deliveryService, log)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router.New(deliveriesHandler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Info("service_started", map[string]any{"service": "service_hub_notifier", "port": port})
	serverErrCh := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdown)

	select {
	case sig := <-shutdown:
		log.Info("shutdown_signal_received", map[string]any{"signal": sig.String()})
	case err := <-serverErrCh:
		log.Error("server_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	log.Info("service_stopped", map[string]any{"service": "service_hub_notifier"})
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func timeoutFromEnv(key string, fallbackMS int) time.Duration {
	raw := envOrDefault(key, fmt.Sprintf("%d", fallbackMS))
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		parsed = fallbackMS
	}

	return time.Duration(parsed) * time.Millisecond
}
