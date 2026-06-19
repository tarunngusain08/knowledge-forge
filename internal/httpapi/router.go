package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/chat"
	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/codeqa"
	"github.com/tarunngusain08/knowledge-forge/internal/config"
	"github.com/tarunngusain08/knowledge-forge/internal/documents"
	"github.com/tarunngusain08/knowledge-forge/internal/evaluation"
	"github.com/tarunngusain08/knowledge-forge/internal/indexing"
	"github.com/tarunngusain08/knowledge-forge/internal/observability"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
	"github.com/tarunngusain08/knowledge-forge/internal/worker"
)

const maxJSONBodyBytes int64 = 1 << 20

type Dependencies struct {
	Config            config.Config
	Logger            *slog.Logger
	Auth              *auth.Service
	Documents         *documents.Service
	Worker            *worker.Service
	Chat              *chat.Service
	Retriever         rag.Retriever
	Evaluation        *evaluation.Service
	Repositories      *repositories.Service
	RepositoryStore   *repositories.Store
	RepositoryIndexer *indexing.RepositoryIndexer
	CodeQA            *codeqa.Service
}

type repositoryTraceStore interface {
	GetRetrievalTraceForUser(ctx context.Context, userID, id uuid.UUID) (repositories.RetrievalTrace, error)
	CreateFeedback(ctx context.Context, input repositories.CreateFeedbackInput) (repositories.Feedback, error)
}

type repositoryService interface {
	Create(ctx context.Context, input repositories.CreateInput) (codeintel.Repository, error)
	Get(ctx context.Context, ownerUserID, repositoryID uuid.UUID) (codeintel.Repository, error)
	CreateIngestion(ctx context.Context, input repositories.CreateIngestionInput) (codeintel.IngestionJob, error)
	GetIngestionForUser(ctx context.Context, userID, jobID uuid.UUID) (codeintel.IngestionJob, error)
}

func NewRouter(deps Dependencies) http.Handler {
	server := &Server{
		auth:                deps.Auth,
		documents:           deps.Documents,
		worker:              deps.Worker,
		chat:                deps.Chat,
		retriever:           deps.Retriever,
		evaluation:          deps.Evaluation,
		repositories:        deps.Repositories,
		repoStore:           deps.RepositoryStore,
		repoIndexer:         deps.RepositoryIndexer,
		codeQA:              deps.CodeQA,
		maxUploadBytes:      deps.Config.MaxUploadBytes,
		internalWorkerToken: deps.Config.InternalWorkerToken,
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(observability.HTTPMiddleware(deps.Logger))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": deps.Config.ServiceName,
		})
	})

	if deps.Auth != nil {
		r.Post("/auth/login", server.handleLogin)
		r.Group(func(protected chi.Router) {
			protected.Use(auth.Middleware(deps.Auth))
			protected.Get("/me", server.handleMe)
			if deps.Documents != nil {
				protected.Post("/documents", server.handleUploadDocument)
				protected.Get("/documents", server.handleListDocuments)
				protected.Get("/documents/{id}", server.handleGetDocument)
				protected.Delete("/documents/{id}", server.handleDeleteDocument)
			}
			if deps.Chat != nil {
				protected.Post("/chat/sessions", server.handleCreateChatSession)
				protected.Get("/chat/sessions/{id}", server.handleGetChatSession)
				protected.Post("/chat/sessions/{id}/messages", server.handleCreateChatMessage)
			}
			if deps.Retriever != nil {
				protected.Get("/debug/retrieval", server.handleDebugRetrieval)
			}
			if deps.Evaluation != nil {
				protected.Post("/eval/runs", server.handleCreateEvalRun)
				protected.Get("/eval/runs/{id}", server.handleGetEvalRun)
			}
			if deps.Repositories != nil {
				protected.Post("/v1/repositories", server.handleCreateRepository)
				protected.Get("/v1/repositories/{repository_id}", server.handleGetRepository)
				protected.Post("/v1/repositories/{repository_id}/ingestions", server.handleCreateRepositoryIngestion)
				protected.Get("/v1/ingestions/{job_id}", server.handleGetRepositoryIngestion)
			}
			if deps.CodeQA != nil {
				protected.Post("/v1/ask", server.handleRepositoryAsk)
				protected.Post("/v1/plans", server.handleGenerateImplementationPlan)
				protected.Post("/v1/impact", server.handleAnalyzeImpact)
				protected.Post("/v1/reports/deep-dive", server.handleGenerateDeepDiveReport)
			}
			if deps.RepositoryStore != nil {
				protected.Get("/v1/retrieval-traces/{trace_id}", server.handleGetRepositoryRetrievalTrace)
				protected.Post("/v1/feedback", server.handleCreateRepositoryFeedback)
			}
		})
	}
	if deps.Worker != nil {
		r.With(server.requireInternalWorkerToken).Post("/internal/jobs/{job_id}/process", server.handleProcessJob)
	}
	if deps.RepositoryIndexer != nil {
		r.With(server.requireInternalWorkerToken).Post("/internal/repository-jobs/{job_id}/process", server.handleProcessRepositoryJob)
	}

	return r
}

type Server struct {
	auth                *auth.Service
	documents           *documents.Service
	worker              *worker.Service
	chat                *chat.Service
	retriever           rag.Retriever
	evaluation          *evaluation.Service
	repositories        repositoryService
	repoStore           repositoryTraceStore
	repoIndexer         *indexing.RepositoryIndexer
	codeQA              *codeqa.Service
	maxUploadBytes      int64
	internalWorkerToken string
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return decodeJSON(w, r, dst, false)
}

func readOptionalJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return decodeJSON(w, r, dst, true)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, allowEmpty bool) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) && allowEmpty {
			return nil
		}
		return fmt.Errorf("decode JSON body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode JSON body: multiple JSON values are not allowed")
	}
	return nil
}
