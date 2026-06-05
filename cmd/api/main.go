package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/chat"
	"github.com/tarunngusain08/knowledge-forge/internal/config"
	"github.com/tarunngusain08/knowledge-forge/internal/database"
	"github.com/tarunngusain08/knowledge-forge/internal/db"
	"github.com/tarunngusain08/knowledge-forge/internal/documents"
	"github.com/tarunngusain08/knowledge-forge/internal/evaluation"
	"github.com/tarunngusain08/knowledge-forge/internal/httpapi"
	"github.com/tarunngusain08/knowledge-forge/internal/observability"
	"github.com/tarunngusain08/knowledge-forge/internal/providers"
	"github.com/tarunngusain08/knowledge-forge/internal/retrieval"
	"github.com/tarunngusain08/knowledge-forge/internal/worker"
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
	shutdownTelemetry, err := observability.Init(ctx, cfg.ServiceName, cfg.Environment)
	if err != nil {
		logger.Error("init telemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(shutdownCtx); err != nil {
			logger.Error("shutdown telemetry", "error", err)
		}
	}()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	authService, err := auth.NewService(db.New(pool), cfg.JWTSecret)
	if err != nil {
		logger.Error("init auth", "error", err)
		os.Exit(1)
	}
	if err := authService.SeedAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		logger.Error("seed admin", "error", err)
		os.Exit(1)
	}
	documentService := documents.NewService(db.New(pool), documents.NoopVirusScanner{}, cfg.MaxUploadBytes)
	indexingProviders, err := providers.NewIndexingProviders(ctx, cfg)
	if err != nil {
		logger.Error("init indexing providers", "error", err)
		os.Exit(1)
	}
	workerService := worker.NewService(
		db.New(pool),
		indexingProviders.Extractor,
		indexingProviders.Chunker,
		indexingProviders.Embedder,
		indexingProviders.Vector,
		logger,
		cfg.WorkerName,
	)
	queryProviders, err := providers.NewQueryProviders(ctx, cfg)
	if err != nil {
		logger.Error("init query providers", "error", err)
		os.Exit(1)
	}
	lexical := retrieval.NewPostgresFTS(db.New(pool))
	retriever := retrieval.NewService(db.New(pool), queryProviders.Embedder, queryProviders.Vector, lexical, queryProviders.Reranker)
	chatService := chat.NewService(db.New(pool), queryProviders.LLM, retriever)
	evalService := evaluation.NewService(db.New(pool), retriever)

	router := httpapi.NewRouter(httpapi.Dependencies{
		Config:     cfg,
		Logger:     logger,
		Auth:       authService,
		Documents:  documentService,
		Worker:     workerService,
		Chat:       chatService,
		Retriever:  retriever,
		Evaluation: evalService,
	})

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown api", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api failed", "error", err)
			os.Exit(1)
		}
	}
}
