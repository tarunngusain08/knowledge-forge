package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/tarunngusain08/knowledge-forge/internal/db"
	"github.com/tarunngusain08/knowledge-forge/internal/documents"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

type Store interface {
	LeaseIndexingJobs(ctx context.Context, arg db.LeaseIndexingJobsParams) ([]db.IndexingJob, error)
	MarkIndexingJobRunning(ctx context.Context, arg db.MarkIndexingJobRunningParams) (db.IndexingJob, error)
	MarkIndexingJobSucceeded(ctx context.Context, id uuid.UUID) (db.IndexingJob, error)
	MarkIndexingJobFailed(ctx context.Context, arg db.MarkIndexingJobFailedParams) (db.IndexingJob, error)
	GetDocumentBytes(ctx context.Context, id uuid.UUID) (db.Document, error)
	MarkDocumentStatus(ctx context.Context, arg db.MarkDocumentStatusParams) (db.MarkDocumentStatusRow, error)
	DeleteChunksByDocument(ctx context.Context, documentID uuid.UUID) error
	CreateChunk(ctx context.Context, arg db.CreateChunkParams) (db.CreateChunkRow, error)
}

type Service struct {
	store      Store
	extractor  rag.DocumentExtractor
	chunker    rag.ChunkingProvider
	embedder   rag.EmbeddingProvider
	vector     rag.VectorStoreProvider
	logger     *slog.Logger
	workerName string
}

func NewService(store Store, extractor rag.DocumentExtractor, chunker rag.ChunkingProvider, embedder rag.EmbeddingProvider, vector rag.VectorStoreProvider, logger *slog.Logger, workerName string) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if workerName == "" {
		workerName = "worker"
	}
	return &Service{store: store, extractor: extractor, chunker: chunker, embedder: embedder, vector: vector, logger: logger, workerName: workerName}
}

func (s *Service) Lease(ctx context.Context, limit int32) ([]db.IndexingJob, error) {
	if limit <= 0 {
		limit = 1
	}
	return s.store.LeaseIndexingJobs(ctx, db.LeaseIndexingJobsParams{
		Limit:    limit,
		LockedBy: pgtype.Text{String: s.workerName, Valid: true},
	})
}

func (s *Service) ProcessJob(ctx context.Context, jobID uuid.UUID) error {
	ctx, span := otel.Tracer("knowledge-forge/worker").Start(ctx, "worker.process_job")
	defer span.End()
	span.SetAttributes(attribute.String("job.id", jobID.String()))
	job, err := s.store.MarkIndexingJobRunning(ctx, db.MarkIndexingJobRunningParams{
		ID:       jobID,
		LockedBy: pgtype.Text{String: s.workerName, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("mark job running: %w", err)
	}
	if err := s.processDocument(ctx, job.DocumentID); err != nil {
		_, _ = s.store.MarkDocumentStatus(ctx, db.MarkDocumentStatusParams{
			ID:           job.DocumentID,
			Status:       documents.StatusFailed,
			ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
		})
		_, _ = s.store.MarkIndexingJobFailed(ctx, db.MarkIndexingJobFailedParams{
			ID:           job.ID,
			ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
		})
		return err
	}
	if _, err := s.store.MarkDocumentStatus(ctx, db.MarkDocumentStatusParams{
		ID:           job.DocumentID,
		Status:       documents.StatusIndexed,
		ErrorMessage: pgtype.Text{Valid: false},
	}); err != nil {
		return fmt.Errorf("mark document indexed: %w", err)
	}
	if _, err := s.store.MarkIndexingJobSucceeded(ctx, job.ID); err != nil {
		return fmt.Errorf("mark job succeeded: %w", err)
	}
	return nil
}

func (s *Service) processDocument(ctx context.Context, documentID uuid.UUID) error {
	document, err := s.store.GetDocumentBytes(ctx, documentID)
	if err != nil {
		return fmt.Errorf("get document bytes: %w", err)
	}
	if _, err := s.store.MarkDocumentStatus(ctx, db.MarkDocumentStatusParams{
		ID:           document.ID,
		Status:       documents.StatusIndexing,
		ErrorMessage: pgtype.Text{Valid: false},
	}); err != nil {
		return fmt.Errorf("mark document indexing: %w", err)
	}

	text, err := s.extractor.Extract(ctx, document.Filename, document.RawBytes)
	if err != nil {
		return fmt.Errorf("extract document text: %w", err)
	}
	chunks, err := s.chunker.Split(ctx, rag.ChunkInput{
		DocumentID: document.ID,
		Filename:   document.Filename,
		Content:    text,
		Metadata: map[string]any{
			"filename":     document.Filename,
			"document_id":  document.ID.String(),
			"content_type": document.ContentType,
			"sha256":       document.Sha256,
		},
	})
	if err != nil {
		return fmt.Errorf("chunk document: %w", err)
	}
	if len(chunks) == 0 {
		return fmt.Errorf("document produced no chunks")
	}
	embeddings, err := s.embedder.EmbedDocuments(ctx, chunkTexts(chunks))
	if err != nil {
		return fmt.Errorf("embed chunks: %w", err)
	}
	if len(embeddings) != len(chunks) {
		return fmt.Errorf("embedding count mismatch: got %d want %d", len(embeddings), len(chunks))
	}

	if err := s.store.DeleteChunksByDocument(ctx, document.ID); err != nil {
		return fmt.Errorf("delete old chunks: %w", err)
	}
	records := make([]rag.VectorRecord, 0, len(chunks))
	for i, chunk := range chunks {
		metadata := chunk.Metadata
		metadata["vector_id"] = chunk.VectorID
		metadata["chunk_index"] = chunk.Index
		metadata["document_id"] = document.ID.String()
		metadataBytes, err := documents.MetadataJSON(metadata)
		if err != nil {
			return fmt.Errorf("marshal chunk metadata: %w", err)
		}
		dbChunk, err := s.store.CreateChunk(ctx, db.CreateChunkParams{
			DocumentID: document.ID,
			ChunkIndex: int32(chunk.Index),
			Content:    chunk.Content,
			PageNumber: pgtype.Int4{Valid: false},
			TokenCount: int32(embeddings[i].InputTokens),
			Metadata:   metadataBytes,
		})
		if err != nil {
			return fmt.Errorf("create chunk: %w", err)
		}
		metadata["chunk_id"] = dbChunk.ID.String()
		records = append(records, rag.VectorRecord{
			ID:       chunk.VectorID,
			Values:   embeddings[i].Vector,
			Metadata: metadata,
		})
	}
	if err := s.vector.UpsertChunks(ctx, records); err != nil {
		return fmt.Errorf("upsert vectors: %w", err)
	}
	s.logger.Info("indexed document", "document_id", document.ID, "chunks", len(chunks))
	return nil
}

func chunkTexts(chunks []rag.Chunk) []string {
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = strings.TrimSpace(chunk.Content)
	}
	return texts
}

func MarshalJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}
