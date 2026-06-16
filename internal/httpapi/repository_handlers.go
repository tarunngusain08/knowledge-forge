package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/codeqa"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
)

type createRepositoryRequest struct {
	Name          string `json:"name"`
	RemoteURL     string `json:"remote_url"`
	LocalPath     string `json:"local_path"`
	DefaultBranch string `json:"default_branch"`
}

func (s *Server) handleCreateRepository(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	var req createRepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	repo, err := s.repositories.Create(r.Context(), repositories.CreateInput{
		OwnerUserID:   user.ID,
		Name:          req.Name,
		RemoteURL:     req.RemoteURL,
		LocalPath:     req.LocalPath,
		DefaultBranch: req.DefaultBranch,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, repo)
}

func (s *Server) handleGetRepository(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	repositoryID, err := uuid.Parse(chi.URLParam(r, "repository_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repository id")
		return
	}
	repo, err := s.repositories.Get(r.Context(), user.ID, repositoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}
	writeJSON(w, http.StatusOK, repo)
}

type createRepositoryIngestionRequest struct {
	BranchName string `json:"branch_name"`
	CommitSHA  string `json:"commit_sha"`
	ProcessNow bool   `json:"process_now"`
}

func (s *Server) handleCreateRepositoryIngestion(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	repositoryID, err := uuid.Parse(chi.URLParam(r, "repository_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repository id")
		return
	}
	var req createRepositoryIngestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	job, err := s.repositories.CreateIngestion(r.Context(), repositories.CreateIngestionInput{
		UserID:       user.ID,
		RepositoryID: repositoryID,
		BranchName:   req.BranchName,
		CommitSHA:    req.CommitSHA,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.ProcessNow {
		if s.repoIndexer == nil {
			writeError(w, http.StatusServiceUnavailable, "repository indexer is not configured")
			return
		}
		if err := s.repoIndexer.ProcessJob(r.Context(), job.ID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		job, _ = s.repositories.GetIngestion(r.Context(), job.ID)
	}
	writeJSON(w, http.StatusAccepted, job)
}

func (s *Server) handleGetRepositoryIngestion(w http.ResponseWriter, r *http.Request) {
	jobID, err := uuid.Parse(chi.URLParam(r, "job_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}
	job, err := s.repositories.GetIngestion(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "ingestion job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleProcessRepositoryJob(w http.ResponseWriter, r *http.Request) {
	if s.repoIndexer == nil {
		writeError(w, http.StatusServiceUnavailable, "repository indexer is not configured")
		return
	}
	jobID, err := uuid.Parse(chi.URLParam(r, "job_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}
	if err := s.repoIndexer.ProcessJob(r.Context(), jobID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

type repositoryAskRequest struct {
	RepositoryID    uuid.UUID `json:"repository_id"`
	BranchName      string    `json:"branch_name"`
	Question        string    `json:"question"`
	TopK            int       `json:"top_k"`
	RerankerEnabled *bool     `json:"reranker_enabled,omitempty"`
}

func (s *Server) handleRepositoryAsk(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing user")
		return
	}
	var req repositoryAskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}
	resp, err := s.codeQA.Ask(r.Context(), codeqa.AskRequest{
		UserID:          user.ID,
		RepositoryID:    req.RepositoryID,
		BranchName:      req.BranchName,
		Question:        req.Question,
		TopK:            req.TopK,
		RerankerEnabled: req.RerankerEnabled,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetRepositoryRetrievalTrace(w http.ResponseWriter, r *http.Request) {
	traceID, err := uuid.Parse(chi.URLParam(r, "trace_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid trace id")
		return
	}
	trace, err := s.repoStore.GetRetrievalTrace(r.Context(), traceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "retrieval trace not found")
		return
	}
	writeJSON(w, http.StatusOK, trace)
}
