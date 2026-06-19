package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
)

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

type traceStoreFake struct {
	ownerID    uuid.UUID
	traceID    uuid.UUID
	trace      repositories.RetrievalTrace
	lastUserID uuid.UUID
}

func (s *traceStoreFake) GetRetrievalTraceForUser(_ context.Context, userID, id uuid.UUID) (repositories.RetrievalTrace, error) {
	s.lastUserID = userID
	if userID != s.ownerID || id != s.traceID {
		return repositories.RetrievalTrace{}, errors.New("not found")
	}
	return s.trace, nil
}

func (s *traceStoreFake) CreateFeedback(context.Context, repositories.CreateFeedbackInput) (repositories.Feedback, error) {
	return repositories.Feedback{}, nil
}
