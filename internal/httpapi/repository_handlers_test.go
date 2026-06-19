package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
)

func TestGetRepositoryIngestionRequiresOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	jobID := uuid.New()
	store := &repositoryServiceFake{
		ownerID: ownerID,
		jobID:   jobID,
		job: codeintel.IngestionJob{
			ID: jobID,
		},
	}
	server := &Server{repositories: store}
	router := chi.NewRouter()
	router.Get("/v1/ingestions/{job_id}", server.handleGetRepositoryIngestion)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingestions/"+jobID.String(), nil)
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: otherUserID, Email: "other@example.com"}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("cross-user ingestion status = %d", recorder.Code)
	}
	if store.lastUserID != otherUserID {
		t.Fatalf("ingestion lookup user = %s, want %s", store.lastUserID, otherUserID)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/ingestions/"+jobID.String(), nil)
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: ownerID, Email: "owner@example.com"}))
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("owner ingestion status = %d", recorder.Code)
	}
}

func TestGetRepositoryRetrievalTraceRequiresOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	traceID := uuid.New()
	store := &traceStoreFake{
		ownerID: ownerID,
		traceID: traceID,
		trace: repositories.RetrievalTrace{
			ID: traceID,
		},
	}
	server := &Server{repoStore: store}
	router := chi.NewRouter()
	router.Get("/v1/retrieval-traces/{trace_id}", server.handleGetRepositoryRetrievalTrace)

	req := httptest.NewRequest(http.MethodGet, "/v1/retrieval-traces/"+traceID.String(), nil)
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: otherUserID, Email: "other@example.com"}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("cross-user trace status = %d", recorder.Code)
	}
	if store.lastUserID != otherUserID {
		t.Fatalf("trace lookup user = %s, want %s", store.lastUserID, otherUserID)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/retrieval-traces/"+traceID.String(), nil)
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: ownerID, Email: "owner@example.com"}))
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("owner trace status = %d", recorder.Code)
	}
}

func TestCreateRepositoryFeedbackRequiresTraceOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	traceID := uuid.New()
	store := &traceStoreFake{
		ownerID: ownerID,
		traceID: traceID,
		trace: repositories.RetrievalTrace{
			ID: traceID,
		},
	}
	server := &Server{repoStore: store}
	router := chi.NewRouter()
	router.Post("/v1/feedback", server.handleCreateRepositoryFeedback)
	body, err := json.Marshal(repositoryFeedbackRequest{TraceID: traceID, AnswerCorrect: true})
	if err != nil {
		t.Fatalf("marshal feedback: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: otherUserID, Email: "other@example.com"}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("cross-user feedback status = %d", recorder.Code)
	}
	if store.createCalled {
		t.Fatalf("feedback insert should not run for a foreign trace")
	}

	req = httptest.NewRequest(http.MethodPost, "/v1/feedback", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: ownerID, Email: "owner@example.com"}))
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("owner feedback status = %d", recorder.Code)
	}
	if !store.createCalled {
		t.Fatalf("feedback insert did not run for owner trace")
	}
}

type repositoryServiceFake struct {
	ownerID    uuid.UUID
	jobID      uuid.UUID
	job        codeintel.IngestionJob
	lastUserID uuid.UUID
}

func (s *repositoryServiceFake) Create(context.Context, repositories.CreateInput) (codeintel.Repository, error) {
	return codeintel.Repository{}, nil
}

func (s *repositoryServiceFake) Get(context.Context, uuid.UUID, uuid.UUID) (codeintel.Repository, error) {
	return codeintel.Repository{}, nil
}

func (s *repositoryServiceFake) CreateIngestion(context.Context, repositories.CreateIngestionInput) (codeintel.IngestionJob, error) {
	return codeintel.IngestionJob{}, nil
}

func (s *repositoryServiceFake) GetIngestionForUser(_ context.Context, userID, jobID uuid.UUID) (codeintel.IngestionJob, error) {
	s.lastUserID = userID
	if userID != s.ownerID || jobID != s.jobID {
		return codeintel.IngestionJob{}, errors.New("not found")
	}
	return s.job, nil
}

type traceStoreFake struct {
	ownerID      uuid.UUID
	traceID      uuid.UUID
	trace        repositories.RetrievalTrace
	lastUserID   uuid.UUID
	createCalled bool
}

func (s *traceStoreFake) GetRetrievalTraceForUser(_ context.Context, userID, id uuid.UUID) (repositories.RetrievalTrace, error) {
	s.lastUserID = userID
	if userID != s.ownerID || id != s.traceID {
		return repositories.RetrievalTrace{}, errors.New("not found")
	}
	return s.trace, nil
}

func (s *traceStoreFake) CreateFeedback(context.Context, repositories.CreateFeedbackInput) (repositories.Feedback, error) {
	s.createCalled = true
	return repositories.Feedback{ID: uuid.New(), TraceID: s.traceID}, nil
}
