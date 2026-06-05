package documents

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/db"
)

const (
	StatusUploaded = "uploaded"
	StatusIndexing = "indexing"
	StatusIndexed  = "indexed"
	StatusFailed   = "failed"
	StatusDeleted  = "deleted"
)

type Store interface {
	CreateDocument(ctx context.Context, arg db.CreateDocumentParams) (db.CreateDocumentRow, error)
	GetDocumentByHash(ctx context.Context, arg db.GetDocumentByHashParams) (db.GetDocumentByHashRow, error)
	ListDocumentsByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]db.ListDocumentsByOwnerRow, error)
	GetDocumentByIDAndOwner(ctx context.Context, arg db.GetDocumentByIDAndOwnerParams) (db.GetDocumentByIDAndOwnerRow, error)
	MarkDocumentDeleted(ctx context.Context, arg db.MarkDocumentDeletedParams) error
	CreateIndexingJob(ctx context.Context, documentID uuid.UUID) (db.IndexingJob, error)
}

type VirusScanner interface {
	Scan(ctx context.Context, filename string, content []byte) error
}

type NoopVirusScanner struct{}

func (NoopVirusScanner) Scan(context.Context, string, []byte) error {
	return nil
}

type Service struct {
	store          Store
	scanner        VirusScanner
	maxUploadBytes int64
}

type UploadInput struct {
	OwnerUserID uuid.UUID
	Filename    string
	Content     []byte
}

type UploadResult struct {
	Document db.CreateDocumentRow `json:"document"`
	Job      db.IndexingJob       `json:"job"`
}

func NewService(store Store, scanner VirusScanner, maxUploadBytes int64) *Service {
	if scanner == nil {
		scanner = NoopVirusScanner{}
	}
	return &Service{store: store, scanner: scanner, maxUploadBytes: maxUploadBytes}
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (UploadResult, error) {
	if input.OwnerUserID == uuid.Nil {
		return UploadResult{}, errors.New("owner user id is required")
	}
	if err := validateFilename(input.Filename); err != nil {
		return UploadResult{}, err
	}
	if s.maxUploadBytes > 0 && int64(len(input.Content)) > s.maxUploadBytes {
		return UploadResult{}, fmt.Errorf("file exceeds %d byte upload limit", s.maxUploadBytes)
	}
	contentType, err := validateContent(input.Filename, input.Content)
	if err != nil {
		return UploadResult{}, err
	}
	if err := s.scanner.Scan(ctx, input.Filename, input.Content); err != nil {
		return UploadResult{}, fmt.Errorf("virus scan failed: %w", err)
	}
	hash := sha256.Sum256(input.Content)
	sha := hex.EncodeToString(hash[:])

	if _, err := s.store.GetDocumentByHash(ctx, db.GetDocumentByHashParams{
		OwnerUserID: input.OwnerUserID,
		Sha256:      sha,
	}); err == nil {
		return UploadResult{}, ErrDuplicateDocument
	}

	document, err := s.store.CreateDocument(ctx, db.CreateDocumentParams{
		OwnerUserID: input.OwnerUserID,
		Filename:    filepath.Base(input.Filename),
		ContentType: contentType,
		SizeBytes:   int64(len(input.Content)),
		Sha256:      sha,
		RawBytes:    input.Content,
	})
	if err != nil {
		return UploadResult{}, fmt.Errorf("create document: %w", err)
	}
	job, err := s.store.CreateIndexingJob(ctx, document.ID)
	if err != nil {
		return UploadResult{}, fmt.Errorf("create indexing job: %w", err)
	}
	return UploadResult{Document: document, Job: job}, nil
}

func (s *Service) List(ctx context.Context, owner uuid.UUID) ([]db.ListDocumentsByOwnerRow, error) {
	return s.store.ListDocumentsByOwner(ctx, owner)
}

func (s *Service) Get(ctx context.Context, owner, id uuid.UUID) (db.GetDocumentByIDAndOwnerRow, error) {
	return s.store.GetDocumentByIDAndOwner(ctx, db.GetDocumentByIDAndOwnerParams{ID: id, OwnerUserID: owner})
}

func (s *Service) Delete(ctx context.Context, owner, id uuid.UUID) error {
	return s.store.MarkDocumentDeleted(ctx, db.MarkDocumentDeletedParams{ID: id, OwnerUserID: owner})
}

func ExtractPlainText(filename string, content []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".md":
		return string(content), nil
	case ".pdf":
		return "", errors.New("PDF extraction is handled by the indexing worker")
	default:
		return "", fmt.Errorf("unsupported file extension %q", ext)
	}
}

func MetadataJSON(metadata map[string]any) ([]byte, error) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	return json.Marshal(metadata)
}

func ReadUpload(r io.Reader, maxBytes int64) ([]byte, error) {
	var buf bytes.Buffer
	limit := maxBytes + 1
	if limit <= 1 {
		limit = 20*1024*1024 + 1
	}
	if _, err := io.CopyN(&buf, r, limit); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if int64(buf.Len()) >= limit {
		return nil, fmt.Errorf("file exceeds %d byte upload limit", maxBytes)
	}
	return buf.Bytes(), nil
}

var ErrDuplicateDocument = errors.New("document already uploaded")

func validateFilename(filename string) error {
	if strings.TrimSpace(filename) == "" {
		return errors.New("filename is required")
	}
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".pdf", ".txt", ".md":
		return nil
	default:
		return errors.New("only PDF, TXT, and Markdown files are supported")
	}
}

func validateContent(filename string, content []byte) (string, error) {
	if len(content) == 0 {
		return "", errors.New("file is empty")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	detected := http.DetectContentType(content)
	switch ext {
	case ".pdf":
		if detected != "application/pdf" {
			return "", fmt.Errorf("expected PDF content, detected %s", detected)
		}
		return detected, nil
	case ".txt":
		if !strings.HasPrefix(detected, "text/plain") {
			return "", fmt.Errorf("expected text content, detected %s", detected)
		}
		return "text/plain; charset=utf-8", nil
	case ".md":
		if !strings.HasPrefix(detected, "text/plain") {
			return "", fmt.Errorf("expected markdown text content, detected %s", detected)
		}
		return "text/markdown; charset=utf-8", nil
	default:
		return "", errors.New("unsupported file extension")
	}
}
