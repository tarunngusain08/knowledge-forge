package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tarunngusain08/knowledge-forge/internal/config"
	"github.com/tarunngusain08/knowledge-forge/internal/database"
	"github.com/tarunngusain08/knowledge-forge/internal/db"
	"github.com/tarunngusain08/knowledge-forge/internal/indexing"
	"github.com/tarunngusain08/knowledge-forge/internal/observability"
	"github.com/tarunngusain08/knowledge-forge/internal/providers"
	gitprovider "github.com/tarunngusain08/knowledge-forge/internal/providers/git"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
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
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	indexingProviders, err := providers.NewIndexingProviders(ctx, cfg)
	if err != nil {
		logger.Error("init indexing providers", "error", err)
		os.Exit(1)
	}
	service := worker.NewService(
		db.New(pool),
		indexingProviders.Extractor,
		indexingProviders.Chunker,
		indexingProviders.Embedder,
		indexingProviders.Vector,
		logger,
		cfg.WorkerName,
	)
	repoStore := repositories.NewStore(pool)
	repoIndexer := indexing.NewRepositoryIndexer(
		repoStore,
		gitprovider.Provider{},
		indexingProviders.Embedder,
		indexingProviders.Vector,
		logger,
		cfg.WorkerName,
	)

	logger.Info("worker started", "service", cfg.ServiceName, "name", cfg.WorkerName)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if err := processBatch(ctx, service, int32(cfg.WorkerBatchSize), logger); err != nil {
			logger.Error("process indexing batch", "error", err)
		}
		if err := processRepositoryBatch(ctx, repoIndexer, int32(cfg.WorkerBatchSize), logger); err != nil {
			logger.Error("process repository indexing batch", "error", err)
		}
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return
		case <-ticker.C:
		}
	}
}

func processBatch(ctx context.Context, service *worker.Service, batchSize int32, logger *slog.Logger) error {
	jobs, err := service.Lease(ctx, batchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := service.ProcessJob(ctx, job.ID); err != nil {
			logger.Error("process job", "job_id", job.ID, "document_id", job.DocumentID, "error", err)
			continue
		}
		logger.Info("processed job", "job_id", job.ID, "document_id", job.DocumentID)
	}
	return nil
}

func processRepositoryBatch(ctx context.Context, service *indexing.RepositoryIndexer, batchSize int32, logger *slog.Logger) error {
	jobs, err := service.Lease(ctx, batchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := service.ProcessJob(ctx, job.ID); err != nil {
			logger.Error("process repository job", "job_id", job.ID, "repository_id", job.RepositoryID, "error", err)
			continue
		}
		logger.Info("processed repository job", "job_id", job.ID, "repository_id", job.RepositoryID)
	}
	return nil
}
