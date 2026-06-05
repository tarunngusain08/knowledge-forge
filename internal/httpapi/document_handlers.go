package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
	"github.com/tarunngusain08/knowledge-forge/internal/documents"
)

func (s *Server) handleUploadDocument(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	if err := r.ParseMultipartForm(s.maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	content, err := documents.ReadUpload(file, s.maxUploadBytes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.documents.Upload(r.Context(), documents.UploadInput{
		OwnerUserID: user.ID,
		Filename:    header.Filename,
		Content:     content,
	})
	if errors.Is(err, documents.ErrDuplicateDocument) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	rows, err := s.documents.List(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list documents failed")
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}
	row, err := s.documents.Get(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "document not found")
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}
	if err := s.documents.Delete(r.Context(), user.ID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "delete document failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
