package httpapi

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/auth"
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

func TestUploadDocumentRejectsOversizedMultipartBeforeFullParse(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "large.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(bytes.Repeat([]byte("a"), int(multipartFormOverheadBytes)+64)); err != nil {
		t.Fatalf("write multipart body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	reader := &countingReadCloser{Reader: bytes.NewReader(body.Bytes())}
	req := httptest.NewRequest(http.MethodPost, "/documents", reader)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = req.WithContext(auth.ContextWithUser(req.Context(), auth.User{ID: uuid.New(), Email: "owner@example.com"}))
	recorder := httptest.NewRecorder()
	server := &Server{maxUploadBytes: 32}

	server.handleUploadDocument(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
	if reader.bytesRead >= int64(body.Len()) {
		t.Fatalf("multipart parser consumed full body before rejecting: read=%d body=%d", reader.bytesRead, body.Len())
	}
	if reader.bytesRead > server.maxUploadBytes+multipartFormOverheadBytes+4096 {
		t.Fatalf("multipart parser read too far before rejecting: read=%d", reader.bytesRead)
	}
}

type countingReadCloser struct {
	io.Reader
	bytesRead int64
}

func (r *countingReadCloser) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.bytesRead += int64(n)
	return n, err
}

func (r *countingReadCloser) Close() error {
	return nil
}
