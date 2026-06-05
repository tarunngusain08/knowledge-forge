package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/evaluation"
)

type evalRunRequest struct {
	Name            string                `json:"name"`
	TopK            int                   `json:"top_k"`
	RerankerEnabled bool                  `json:"reranker_enabled"`
	Questions       []evaluation.Question `json:"questions"`
}

func (s *Server) handleCreateEvalRun(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	var req evalRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := s.evaluation.Run(r.Context(), evaluation.RunRequest{
		UserID:          user.ID,
		Name:            req.Name,
		TopK:            req.TopK,
		RerankerEnabled: req.RerankerEnabled,
		Questions:       req.Questions,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleGetEvalRun(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid eval run id")
		return
	}
	run, err := s.evaluation.Get(r.Context(), user.ID, runID)
	if err != nil {
		writeError(w, http.StatusNotFound, "eval run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}
