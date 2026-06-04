package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tarunngusain08/RAG-bot/internal/config"
	"github.com/tarunngusain08/RAG-bot/internal/observability"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := observability.NewLogger(cfg.LogLevel)
	logger.Info("worker bootstrapped", "service", cfg.ServiceName)
	<-ctx.Done()
	logger.Info("worker stopped")
}

