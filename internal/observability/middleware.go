package observability

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func HTTPMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	tracer := otel.Tracer("knowledge-forge/http")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
			defer span.End()
			span.SetAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("url.path", r.URL.Path),
			)
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r.WithContext(ctx))
			duration := time.Since(start)
			span.SetAttributes(
				attribute.Int("http.response.status_code", recorder.status),
				attribute.Int64("http.server.duration_ms", duration.Milliseconds()),
			)
			if logger != nil {
				logger.Info("http request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", recorder.status,
					"duration_ms", duration.Milliseconds(),
					"request_id", r.Context().Value("chi.middleware/requestID"),
				)
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
