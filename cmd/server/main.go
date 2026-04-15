package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julianlinaress/service_hub_notifier/internal/adapters/httpclient"
	"github.com/julianlinaress/service_hub_notifier/internal/adapters/logger"
	"github.com/julianlinaress/service_hub_notifier/internal/adapters/providers"
	"github.com/julianlinaress/service_hub_notifier/internal/app"
	"github.com/julianlinaress/service_hub_notifier/internal/config"
	"github.com/julianlinaress/service_hub_notifier/internal/http/handlers"
	"github.com/julianlinaress/service_hub_notifier/internal/http/router"
	"github.com/julianlinaress/service_hub_notifier/internal/service"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdown)

	if err := run(os.Getenv, shutdown); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(getEnv config.EnvGetter, shutdown <-chan os.Signal) error {
	runtimeConfig := config.LoadFromEnv(getEnv)
	log := logger.New()
	readiness := app.NewReadiness(runtimeConfig)

	httpClient := httpclient.New(runtimeConfig.DeliveryTimeout)
	telegram := providers.NewTelegramAdapter(httpClient, providers.WithTelegramAPIBaseURL(runtimeConfig.TelegramAPIBaseURL))
	slack := providers.NewSlackAdapter(httpClient)

	deliveryService := service.NewDeliveryService(telegram, slack)
	deliveriesHandler := handlers.NewDeliveriesHandler(deliveryService, log, runtimeConfig.InternalServiceToken)

	server := &http.Server{
		Addr:              ":" + runtimeConfig.Port,
		Handler:           router.New(deliveriesHandler, readiness),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Info("service_started", map[string]any{"service": "service_hub_notifier", "port": runtimeConfig.Port})
	serverErrCh := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
	}()

	select {
	case sig := <-shutdown:
		log.Info("shutdown_signal_received", map[string]any{"signal": sig.String()})
	case err := <-serverErrCh:
		log.Error("server_failed", map[string]any{"error": err.Error()})
		return fmt.Errorf("run server: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), runtimeConfig.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown_failed", map[string]any{"error": err.Error()})
		return fmt.Errorf("shutdown server: %w", err)
	}

	log.Info("service_stopped", map[string]any{"service": "service_hub_notifier"})

	return nil
}
