package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func (s *Server) requireInternalWorkerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := strings.TrimSpace(s.internalWorkerToken)
		if expected == "" {
			writeError(w, http.StatusServiceUnavailable, "internal worker token is not configured")
			return
		}
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			token = strings.TrimSpace(r.Header.Get("X-Internal-Worker-Token"))
		}
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing internal worker token")
			return
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			writeError(w, http.StatusForbidden, "invalid internal worker token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
