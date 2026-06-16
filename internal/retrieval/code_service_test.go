package retrieval

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

func TestCodeServiceUsesCandidateBudgetAndAnnotatesProvenance(t *testing.T) {
	repositoryID := uuid.New()
	snapshotID := uuid.New()
	chunkID := uuid.New()
	store := &fakeCodeStore{
		snapshot: codeintel.Snapshot{
			ID:           snapshotID,
			RepositoryID: repositoryID,
			BranchName:   "main",
			CommitSHA:    "abc123",
			Status:       "indexed",
			CreatedAt:    time.Now(),
		},
		chunk: codeintel.Chunk{
			ID:           chunkID,
			RepositoryID: repositoryID,
			SnapshotID:   snapshotID,
			FileID:       uuid.New(),
			ChunkIndex:   0,
			ChunkType:    "code",
			Path:         "internal/auth/service.go",
			Language:     "go",
			StartLine:    10,
			EndLine:      18,
			Content:      "func Login() error { return nil }",
			TokenCount:   8,
		},
	}
	vector := &capturingVectorStore{hits: []rag.RetrievalHit{{Chunk: rag.Chunk{VectorID: chunkID.String()}, DenseScore: 0.9}}}
	service := NewCodeService(store, fakeEmbedder{}, vector, nil)

	result, err := service.Retrieve(context.Background(), rag.RetrievalRequest{
		RepositoryID:       repositoryID,
		BranchName:         "main",
		Query:              "where is auth",
		TopK:               5,
		CandidateK:         12,
		QueryCategory:      CategoryExactLookup,
		RetrievalPath:      []string{"dense"},
		RetrievalConfig:    map[string]any{"candidate_k": 12},
		RerankerEnabled:    false,
		ContextTokenBudget: 1200,
	})
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}

	if vector.topK != 12 {
		t.Fatalf("vector topK = %d", vector.topK)
	}
	if result.CommitSHA != "abc123" {
		t.Fatalf("commit sha = %s", result.CommitSHA)
	}
	if got := result.RerankedHits[0].Chunk.Metadata["commit_sha"]; got != "abc123" {
		t.Fatalf("hit commit metadata = %v", got)
	}
	if result.StageContributions["dense"] != 1 {
		t.Fatalf("dense stage contribution = %d", result.StageContributions["dense"])
	}
	if len(result.RetrievedChunkIDs) != 1 || result.RetrievedChunkIDs[0] != chunkID {
		t.Fatalf("retrieved chunk ids = %#v", result.RetrievedChunkIDs)
	}
}

type fakeCodeStore struct {
	snapshot codeintel.Snapshot
	chunk    codeintel.Chunk
}

func (s *fakeCodeStore) LatestSnapshot(context.Context, uuid.UUID, string) (codeintel.Snapshot, error) {
	return s.snapshot, nil
}

func (s *fakeCodeStore) GetChunk(context.Context, uuid.UUID, uuid.UUID) (codeintel.Chunk, error) {
	return s.chunk, nil
}

type fakeEmbedder struct{}

func (fakeEmbedder) EmbedDocuments(context.Context, []string) ([]rag.EmbeddingResult, error) {
	return nil, nil
}

func (fakeEmbedder) EmbedQuery(context.Context, string) (rag.EmbeddingResult, error) {
	return rag.EmbeddingResult{Vector: []float32{0.1, 0.2}, InputTokens: 4, Model: "mock-embedding"}, nil
}

type capturingVectorStore struct {
	topK int
	hits []rag.RetrievalHit
}

func (s *capturingVectorStore) UpsertChunks(context.Context, []rag.VectorRecord) error {
	return nil
}

func (s *capturingVectorStore) Search(_ context.Context, _ []float32, topK int, _ map[string]any) ([]rag.RetrievalHit, error) {
	s.topK = topK
	return s.hits, nil
}

func (s *capturingVectorStore) DeleteDocument(context.Context, uuid.UUID) error {
	return nil
}

func (s *capturingVectorStore) Healthcheck(context.Context) error {
	return nil
}
