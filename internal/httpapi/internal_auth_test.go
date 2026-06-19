package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireInternalWorkerTokenFailsClosedWhenUnset(t *testing.T) {
	server := &Server{}
	called := false
	handler := server.requireInternalWorkerToken(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/internal/jobs/id/process", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", recorder.Code)
	}
	if called {
		t.Fatal("handler should not run when token is not configured")
	}
}

func TestRequireInternalWorkerTokenRejectsMissingOrInvalidToken(t *testing.T) {
	server := &Server{internalWorkerToken: "secret"}
	handler := server.requireInternalWorkerToken(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/internal/jobs/id/process", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d", recorder.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/internal/jobs/id/process", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("invalid token status = %d", recorder.Code)
	}
}

func TestRequireInternalWorkerTokenAcceptsBearerOrHeaderToken(t *testing.T) {
	for _, tt := range []struct {
		name   string
		header string
		value  string
	}{
		{name: "bearer", header: "Authorization", value: "Bearer secret"},
		{name: "internal header", header: "X-Internal-Worker-Token", value: "secret"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{internalWorkerToken: "secret"}
			called := false
			handler := server.requireInternalWorkerToken(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))
			req := httptest.NewRequest(http.MethodPost, "/internal/jobs/id/process", nil)
			req.Header.Set(tt.header, tt.value)

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d", recorder.Code)
			}
			if !called {
				t.Fatal("handler should run with valid token")
			}
		})
	}
}
