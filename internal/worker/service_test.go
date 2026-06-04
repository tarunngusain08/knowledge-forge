package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tarunngusain08/RAG-bot/internal/db"
	"github.com/tarunngusain08/RAG-bot/internal/providers/langchain"
	"github.com/tarunngusain08/RAG-bot/internal/providers/mock"
)

func TestProcessJobIndexesDocument(t *testing.T) {
	documentID := uuid.New()
	jobID := uuid.New()
	store := &fakeStore{
		job: db.IndexingJob{ID: jobID, DocumentID: documentID, Status: "queued"},
		document: db.Document{
			ID:          documentID,
			OwnerUserID: uuid.New(),
			Filename:    "handbook.md",
			ContentType: "text/markdown; charset=utf-8",
			SizeBytes:   128,
			Sha256:      "abc",
			RawBytes:    []byte("# Handbook\nRemote work is allowed with manager approval."),
			Status:      "uploaded",
		},
	}
	vector := &mock.VectorStore{}
	service := NewService(
		store,
		langchain.Extractor{},
		langchain.RecursiveChunker{ChunkSize: 80, ChunkOverlap: 5},
		mock.Embeddings{Dimension: 8},
		vector,
		slog.Default(),
		"test-worker",
	)

	if err := service.ProcessJob(context.Background(), jobID); err != nil {
		t.Fatalf("process job: %v", err)
	}
	if store.documentStatus != "indexed" {
		t.Fatalf("expected indexed status, got %s", store.documentStatus)
	}
	if store.jobStatus != "succeeded" {
		t.Fatalf("expected succeeded job, got %s", store.jobStatus)
	}
	if len(store.chunks) == 0 {
		t.Fatal("expected chunks to be persisted")
	}
	if len(vector.Records) != len(store.chunks) {
		t.Fatalf("expected %d vector records, got %d", len(store.chunks), len(vector.Records))
	}
}

type fakeStore struct {
	job            db.IndexingJob
	document       db.Document
	chunks         []db.CreateChunkRow
	documentStatus string
	jobStatus      string
}

func (f *fakeStore) LeaseIndexingJobs(context.Context, db.LeaseIndexingJobsParams) ([]db.IndexingJob, error) {
	return []db.IndexingJob{f.job}, nil
}

func (f *fakeStore) MarkIndexingJobRunning(context.Context, db.MarkIndexingJobRunningParams) (db.IndexingJob, error) {
	f.job.Status = "running"
	return f.job, nil
}

func (f *fakeStore) MarkIndexingJobSucceeded(context.Context, uuid.UUID) (db.IndexingJob, error) {
	f.job.Status = "succeeded"
	f.jobStatus = "succeeded"
	return f.job, nil
}

func (f *fakeStore) MarkIndexingJobFailed(_ context.Context, _ db.MarkIndexingJobFailedParams) (db.IndexingJob, error) {
	f.job.Status = "failed"
	f.jobStatus = "failed"
	return f.job, nil
}

func (f *fakeStore) GetDocumentBytes(context.Context, uuid.UUID) (db.Document, error) {
	if f.document.ID == uuid.Nil {
		return db.Document{}, errors.New("missing document")
	}
	return f.document, nil
}

func (f *fakeStore) MarkDocumentStatus(_ context.Context, arg db.MarkDocumentStatusParams) (db.MarkDocumentStatusRow, error) {
	f.documentStatus = arg.Status
	return db.MarkDocumentStatusRow{ID: arg.ID, Status: arg.Status}, nil
}

func (f *fakeStore) DeleteChunksByDocument(context.Context, uuid.UUID) error {
	f.chunks = nil
	return nil
}

func (f *fakeStore) CreateChunk(_ context.Context, arg db.CreateChunkParams) (db.CreateChunkRow, error) {
	row := db.CreateChunkRow{
		ID:         uuid.New(),
		DocumentID: arg.DocumentID,
		ChunkIndex: arg.ChunkIndex,
		Content:    arg.Content,
		PageNumber: pgtype.Int4{Valid: false},
		TokenCount: arg.TokenCount,
		Metadata:   arg.Metadata,
	}
	f.chunks = append(f.chunks, row)
	return row, nil
}
