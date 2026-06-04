package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tarunngusain08/RAG-bot/internal/auth"
	"github.com/tarunngusain08/RAG-bot/internal/config"
)

type Dependencies struct {
	Config config.Config
	Logger *slog.Logger
	Auth   *auth.Service
}

func NewRouter(deps Dependencies) http.Handler {
	server := &Server{auth: deps.Auth}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

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
		})
	}

	return r
}

type Server struct {
	auth *auth.Service
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
