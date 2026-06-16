package retrieval

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

type CodeStore interface {
	LatestSnapshot(ctx context.Context, repositoryID uuid.UUID, branchName string) (codeintel.Snapshot, error)
	GetChunk(ctx context.Context, repositoryID uuid.UUID, chunkID uuid.UUID) (codeintel.Chunk, error)
}

type CodeService struct {
	store    CodeStore
	embedder rag.EmbeddingProvider
	vector   rag.VectorStoreProvider
	reranker rag.RerankerProvider
}

func NewCodeService(store CodeStore, embedder rag.EmbeddingProvider, vector rag.VectorStoreProvider, reranker rag.RerankerProvider) *CodeService {
	return &CodeService{store: store, embedder: embedder, vector: vector, reranker: reranker}
}

func (s *CodeService) Retrieve(ctx context.Context, req rag.RetrievalRequest) (rag.RetrievalResult, error) {
	start := time.Now()
	ctx, span := otel.Tracer("knowledge-forge/code-retrieval").Start(ctx, "code_retrieval.retrieve")
	defer span.End()
	if req.RepositoryID == uuid.Nil {
		return rag.RetrievalResult{}, fmt.Errorf("repository id is required")
	}
	branch := strings.TrimSpace(req.BranchName)
	if branch == "" {
		branch = "main"
	}
	topK := req.TopK
	if topK <= 0 {
		topK = 8
	}
	candidateK := req.CandidateK
	if candidateK <= 0 {
		candidateK = topK
	}
	query := strings.TrimSpace(req.Query)
	snapshot, err := s.store.LatestSnapshot(ctx, req.RepositoryID, branch)
	if err != nil {
		return rag.RetrievalResult{}, fmt.Errorf("load latest repository snapshot: %w", err)
	}
	embedding, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return rag.RetrievalResult{}, fmt.Errorf("embed query: %w", err)
	}
	filter := map[string]any{
		"repository_id": map[string]any{"$eq": req.RepositoryID.String()},
		"snapshot_id":   map[string]any{"$eq": snapshot.ID.String()},
	}
	denseRaw, err := s.vector.Search(ctx, embedding.Vector, candidateK, filter)
	if err != nil {
		return rag.RetrievalResult{}, fmt.Errorf("semantic code search: %w", err)
	}
	denseHits, err := s.hydrateDense(ctx, req.RepositoryID, denseRaw)
	if err != nil {
		return rag.RetrievalResult{}, err
	}
	annotateRepositoryHits(denseHits, branch, snapshot.CommitSHA)
	span.SetAttributes(
		attribute.String("repository.id", req.RepositoryID.String()),
		attribute.String("repository.branch", branch),
		attribute.String("repository.snapshot", snapshot.ID.String()),
		attribute.Int("rag.dense_hits", len(denseHits)),
		attribute.Int("rag.candidate_k", candidateK),
		attribute.String("rag.query_category", req.QueryCategory),
		attribute.Bool("rag.reranker_enabled", req.RerankerEnabled),
	)
	reranked := denseHits
	if req.RerankerEnabled && s.reranker != nil && len(denseHits) > 0 {
		reranked, err = s.rerank(ctx, query, denseHits, topK)
		if err != nil {
			return rag.RetrievalResult{}, err
		}
	}
	return rag.RetrievalResult{
		OriginalQuery:     req.Query,
		RewrittenQuery:    query,
		RepositoryID:      req.RepositoryID,
		SnapshotID:        snapshot.ID,
		BranchName:        branch,
		CommitSHA:         snapshot.CommitSHA,
		DenseHits:         denseHits,
		FusedHits:         denseHits,
		RerankedHits:      reranked,
		QueryCategory:     req.QueryCategory,
		RetrievalPath:     req.RetrievalPath,
		RetrievalConfig:   req.RetrievalConfig,
		RetrievedChunkIDs: chunkIDs(reranked),
		StageContributions: map[string]int{
			"dense":  len(denseHits),
			"rerank": rerankContribution(req.RerankerEnabled, reranked),
		},
		Latency: time.Since(start),
	}, nil
}

func annotateRepositoryHits(hits []rag.RetrievalHit, branchName string, commitSHA string) {
	for i := range hits {
		if hits[i].Chunk.Metadata == nil {
			hits[i].Chunk.Metadata = map[string]any{}
		}
		hits[i].Chunk.Metadata["branch_name"] = branchName
		hits[i].Chunk.Metadata["commit_sha"] = commitSHA
	}
}

func (s *CodeService) hydrateDense(ctx context.Context, repositoryID uuid.UUID, hits []rag.RetrievalHit) ([]rag.RetrievalHit, error) {
	hydrated := make([]rag.RetrievalHit, 0, len(hits))
	for _, hit := range hits {
		chunkID, err := uuid.Parse(hit.Chunk.VectorID)
		if err != nil {
			continue
		}
		chunk, err := s.store.GetChunk(ctx, repositoryID, chunkID)
		if err != nil {
			return nil, fmt.Errorf("hydrate code dense hit %s: %w", hit.Chunk.VectorID, err)
		}
		hit.Chunk = ragChunk(chunk)
		hit.Source = "dense"
		hit.Reasons = appendMissing(hit.Reasons, "pinecone")
		hydrated = append(hydrated, hit)
	}
	return hydrated, nil
}

func (s *CodeService) rerank(ctx context.Context, query string, hits []rag.RetrievalHit, topK int) ([]rag.RetrievalHit, error) {
	docs := make([]rag.RerankDocument, 0, len(hits))
	byID := map[string]rag.RetrievalHit{}
	for _, hit := range hits {
		id := hit.Chunk.VectorID
		if id == "" {
			id = hit.Chunk.ID.String()
		}
		docs = append(docs, rag.RerankDocument{ID: id, Content: hit.Chunk.Content})
		byID[id] = hit
	}
	results, err := s.reranker.Rerank(ctx, query, docs, topK)
	if err != nil {
		return nil, fmt.Errorf("rerank code candidates: %w", err)
	}
	reranked := make([]rag.RetrievalHit, 0, len(results))
	for idx, result := range results {
		hit, ok := byID[result.ID]
		if !ok {
			continue
		}
		hit.RerankScore = result.Score
		hit.FusedRank = idx + 1
		hit.Reasons = appendMissing(hit.Reasons, "vertex_ranking")
		reranked = append(reranked, hit)
	}
	return reranked, nil
}

func ragChunk(chunk codeintel.Chunk) rag.Chunk {
	metadata := chunk.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["path"] = chunk.Path
	metadata["language"] = chunk.Language
	metadata["start_line"] = chunk.StartLine
	metadata["end_line"] = chunk.EndLine
	metadata["chunk_type"] = chunk.ChunkType
	metadata["repository_id"] = chunk.RepositoryID.String()
	metadata["snapshot_id"] = chunk.SnapshotID.String()
	return rag.Chunk{
		ID:         chunk.ID,
		DocumentID: chunk.FileID,
		VectorID:   chunk.ID.String(),
		Index:      chunk.ChunkIndex,
		Content:    chunk.Content,
		TokenCount: chunk.TokenCount,
		Metadata:   metadata,
	}
}

func chunkIDs(hits []rag.RetrievalHit) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(hits))
	for _, hit := range hits {
		if hit.Chunk.ID != uuid.Nil {
			ids = append(ids, hit.Chunk.ID)
		}
	}
	return ids
}

func rerankContribution(enabled bool, hits []rag.RetrievalHit) int {
	if !enabled {
		return 0
	}
	return len(hits)
}
