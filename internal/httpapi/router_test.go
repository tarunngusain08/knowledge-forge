package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSONSetsSafeResponseHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeJSON(recorder, http.StatusAccepted, map[string]string{"status": "queued"})

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q", got)
	}
	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("nosniff = %q", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"status":"queued"`) {
		t.Fatalf("body = %q", body)
	}
}

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	var payload struct {
		Name string `json:"name"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"demo","extra":true}`))
	recorder := httptest.NewRecorder()

	if err := readJSON(recorder, req, &payload); err == nil {
		t.Fatalf("expected unknown field to fail")
	}
}

func TestReadJSONRejectsMultipleValues(t *testing.T) {
	var payload struct {
		Name string `json:"name"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"demo"} {"name":"other"}`))
	recorder := httptest.NewRecorder()

	if err := readJSON(recorder, req, &payload); err == nil {
		t.Fatalf("expected multiple JSON values to fail")
	}
}

func TestReadOptionalJSONAllowsEmptyBody(t *testing.T) {
	var payload struct {
		Title string `json:"title"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	recorder := httptest.NewRecorder()

	if err := readOptionalJSON(recorder, req, &payload); err != nil {
		t.Fatalf("read optional JSON: %v", err)
	}
	if payload.Title != "" {
		t.Fatalf("title = %q", payload.Title)
	}
}

func TestReadJSONRejectsOversizedBody(t *testing.T) {
	var payload struct {
		Name string `json:"name"`
	}
	body := `{"name":"` + strings.Repeat("x", int(maxJSONBodyBytes)+1) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	recorder := httptest.NewRecorder()

	if err := readJSON(recorder, req, &payload); err == nil {
		t.Fatalf("expected oversized body to fail")
	}
}
