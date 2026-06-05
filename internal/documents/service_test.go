package documents

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/db"
)

func TestUploadCreatesDocumentAndIndexingJob(t *testing.T) {
	store := &fakeDocumentStore{}
	service := NewService(store, NoopVirusScanner{}, 1024)
	result, err := service.Upload(context.Background(), UploadInput{
		OwnerUserID: uuid.New(),
		Filename:    "handbook.md",
		Content:     []byte("# PTO\nEmployees get time off."),
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if result.Document.Filename != "handbook.md" {
		t.Fatalf("unexpected filename: %s", result.Document.Filename)
	}
	if result.Job.DocumentID != result.Document.ID {
		t.Fatalf("job document mismatch")
	}
}

func TestUploadRejectsDuplicateHash(t *testing.T) {
	store := &fakeDocumentStore{duplicate: true}
	service := NewService(store, NoopVirusScanner{}, 1024)
	_, err := service.Upload(context.Background(), UploadInput{
		OwnerUserID: uuid.New(),
		Filename:    "handbook.txt",
		Content:     []byte("company policy"),
	})
	if !errors.Is(err, ErrDuplicateDocument) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestReadUploadHonorsLimit(t *testing.T) {
	_, err := ReadUpload(strings.NewReader("abcdef"), 5)
	if err == nil {
		t.Fatal("expected upload limit error")
	}
}

type fakeDocumentStore struct {
	duplicate bool
}

func (f *fakeDocumentStore) CreateDocument(_ context.Context, arg db.CreateDocumentParams) (db.CreateDocumentRow, error) {
	return db.CreateDocumentRow{
		ID:          uuid.New(),
		OwnerUserID: arg.OwnerUserID,
		Filename:    arg.Filename,
		ContentType: arg.ContentType,
		SizeBytes:   arg.SizeBytes,
		Sha256:      arg.Sha256,
		Status:      StatusUploaded,
	}, nil
}

func (f *fakeDocumentStore) GetDocumentByHash(context.Context, db.GetDocumentByHashParams) (db.GetDocumentByHashRow, error) {
	if f.duplicate {
		return db.GetDocumentByHashRow{ID: uuid.New()}, nil
	}
	return db.GetDocumentByHashRow{}, errors.New("not found")
}

func (f *fakeDocumentStore) ListDocumentsByOwner(context.Context, uuid.UUID) ([]db.ListDocumentsByOwnerRow, error) {
	return nil, nil
}

func (f *fakeDocumentStore) GetDocumentByIDAndOwner(context.Context, db.GetDocumentByIDAndOwnerParams) (db.GetDocumentByIDAndOwnerRow, error) {
	return db.GetDocumentByIDAndOwnerRow{}, nil
}

func (f *fakeDocumentStore) MarkDocumentDeleted(context.Context, db.MarkDocumentDeletedParams) error {
	return nil
}

func (f *fakeDocumentStore) CreateIndexingJob(_ context.Context, documentID uuid.UUID) (db.IndexingJob, error) {
	return db.IndexingJob{ID: uuid.New(), DocumentID: documentID, Status: "queued"}, nil
}
