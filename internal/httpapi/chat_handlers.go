package httpapi

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/chat"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

type createSessionRequest struct {
	Title string `json:"title"`
}

func (s *Server) handleCreateChatSession(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	var req createSessionRequest
	if err := readOptionalJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	session, err := s.chat.CreateSession(r.Context(), user.ID, req.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleGetChatSession(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	session, messages, err := s.chat.GetSession(r.Context(), user.ID, sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session, "messages": messages})
}

type askRequest struct {
	Question        string `json:"question"`
	TopK            int    `json:"top_k"`
	RerankerEnabled bool   `json:"reranker_enabled"`
}

func (s *Server) handleCreateChatMessage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	var req askRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.TopK == 0 {
		req.TopK = 5
	}
	resp, err := s.chat.Ask(r.Context(), chat.AskRequest{
		UserID:          user.ID,
		SessionID:       sessionID,
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

func (s *Server) handleDebugRetrieval(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	question := r.URL.Query().Get("question")
	if question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}
	topK, _ := strconv.Atoi(r.URL.Query().Get("top_k"))
	if topK == 0 {
		topK = 5
	}
	reranker := r.URL.Query().Get("reranker") == "true"
	result, err := s.retriever.Retrieve(r.Context(), rag.RetrievalRequest{
		UserID:          user.ID,
		Query:           question,
		TopK:            topK,
		RerankerEnabled: reranker,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
