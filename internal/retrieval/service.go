package retrieval

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tarunngusain08/RAG-bot/internal/db"
	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type ChunkStore interface {
	GetChunkByVectorID(ctx context.Context, arg db.GetChunkByVectorIDParams) (db.GetChunkByVectorIDRow, error)
}

type Service struct {
	store    ChunkStore
	embedder rag.EmbeddingProvider
	vector   rag.VectorStoreProvider
	lexical  rag.LexicalSearchProvider
	reranker rag.RerankerProvider
}

func NewService(store ChunkStore, embedder rag.EmbeddingProvider, vector rag.VectorStoreProvider, lexical rag.LexicalSearchProvider, reranker rag.RerankerProvider) *Service {
	return &Service{store: store, embedder: embedder, vector: vector, lexical: lexical, reranker: reranker}
}

func (s *Service) Retrieve(ctx context.Context, req rag.RetrievalRequest) (rag.RetrievalResult, error) {
	start := time.Now()
	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}
	candidateK := 20
	query := strings.TrimSpace(req.Query)
	embedding, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return rag.RetrievalResult{}, fmt.Errorf("embed query: %w", err)
	}
	denseHits, err := s.vector.Search(ctx, embedding.Vector, candidateK, nil)
	if err != nil {
		return rag.RetrievalResult{}, fmt.Errorf("dense search: %w", err)
	}
	hydratedDense, err := s.hydrateDense(ctx, denseHits)
	if err != nil {
		return rag.RetrievalResult{}, err
	}
	lexicalHits, err := s.lexical.Search(ctx, query, candidateK)
	if err != nil {
		return rag.RetrievalResult{}, fmt.Errorf("lexical search: %w", err)
	}
	fused := ReciprocalRankFusion(hydratedDense, lexicalHits, candidateK)
	reranked := fused
	if req.RerankerEnabled && s.reranker != nil && len(fused) > 0 {
		reranked, err = s.rerank(ctx, query, fused, topK)
		if err != nil {
			return rag.RetrievalResult{}, err
		}
	}
	if len(reranked) > topK {
		reranked = reranked[:topK]
	}
	return rag.RetrievalResult{
		OriginalQuery:  req.Query,
		RewrittenQuery: query,
		DenseHits:      hydratedDense,
		LexicalHits:    lexicalHits,
		FusedHits:      fused,
		RerankedHits:   reranked,
		Latency:        time.Since(start),
	}, nil
}

func (s *Service) rerank(ctx context.Context, query string, hits []rag.RetrievalHit, topK int) ([]rag.RetrievalHit, error) {
	docs := make([]rag.RerankDocument, 0, len(hits))
	byID := map[string]rag.RetrievalHit{}
	for _, hit := range hits {
		id := hit.Chunk.VectorID
		docs = append(docs, rag.RerankDocument{ID: id, Content: hit.Chunk.Content})
		byID[id] = hit
	}
	results, err := s.reranker.Rerank(ctx, query, docs, topK)
	if err != nil {
		return nil, fmt.Errorf("rerank candidates: %w", err)
	}
	reranked := make([]rag.RetrievalHit, 0, len(results))
	for i, result := range results {
		hit, ok := byID[result.ID]
		if !ok {
			continue
		}
		hit.RerankScore = result.Score
		hit.FusedRank = i + 1
		hit.Reasons = appendMissing(hit.Reasons, "vertex_ranking")
		reranked = append(reranked, hit)
	}
	return reranked, nil
}

func (s *Service) hydrateDense(ctx context.Context, hits []rag.RetrievalHit) ([]rag.RetrievalHit, error) {
	hydrated := make([]rag.RetrievalHit, 0, len(hits))
	for _, hit := range hits {
		documentID, chunkIndex, ok := parseVectorID(hit.Chunk.VectorID)
		if !ok {
			documentID, chunkIndex, ok = parseMetadata(hit.Chunk.Metadata)
		}
		if !ok {
			continue
		}
		row, err := s.store.GetChunkByVectorID(ctx, db.GetChunkByVectorIDParams{
			DocumentID: documentID,
			ChunkIndex: int32(chunkIndex),
		})
		if err != nil {
			return nil, fmt.Errorf("hydrate dense hit %s: %w", hit.Chunk.VectorID, err)
		}
		hit.Chunk = chunkFromVectorRow(row)
		hit.Source = "dense"
		hit.Reasons = appendMissing(hit.Reasons, "pinecone")
		hydrated = append(hydrated, hit)
	}
	return hydrated, nil
}

func parseVectorID(value string) (uuid.UUID, int, bool) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return uuid.Nil, 0, false
	}
	documentID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, 0, false
	}
	chunkIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return uuid.Nil, 0, false
	}
	return documentID, chunkIndex, true
}

func parseMetadata(metadata map[string]any) (uuid.UUID, int, bool) {
	documentValue, ok := metadata["document_id"].(string)
	if !ok {
		return uuid.Nil, 0, false
	}
	documentID, err := uuid.Parse(documentValue)
	if err != nil {
		return uuid.Nil, 0, false
	}
	switch value := metadata["chunk_index"].(type) {
	case float64:
		return documentID, int(value), true
	case int:
		return documentID, value, true
	default:
		return uuid.Nil, 0, false
	}
}
