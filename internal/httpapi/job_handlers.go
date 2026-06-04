package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) handleProcessJob(w http.ResponseWriter, r *http.Request) {
	if s.worker == nil {
		writeError(w, http.StatusServiceUnavailable, "worker service is not configured")
		return
	}
	jobID, err := uuid.Parse(chi.URLParam(r, "job_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}
	if err := s.worker.ProcessJob(r.Context(), jobID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}
