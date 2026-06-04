package mock

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type Embeddings struct {
	Model      string
	Dimension  int
	LastTask   string
	TokenScale int
}

func (e Embeddings) EmbedDocuments(ctx context.Context, texts []string) ([]rag.EmbeddingResult, error) {
	e.LastTask = rag.TaskRetrievalDocument
	return e.embed(ctx, texts)
}

func (e Embeddings) EmbedQuery(ctx context.Context, text string) (rag.EmbeddingResult, error) {
	results, err := e.embed(ctx, []string{text})
	if err != nil {
		return rag.EmbeddingResult{}, err
	}
	return results[0], nil
}

func (e Embeddings) embed(ctx context.Context, texts []string) ([]rag.EmbeddingResult, error) {
	dim := e.Dimension
	if dim == 0 {
		dim = 16
	}
	model := e.Model
	if model == "" {
		model = "mock-embedding"
	}
	results := make([]rag.EmbeddingResult, 0, len(texts))
	for _, text := range texts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		vector := make([]float32, dim)
		words := strings.Fields(strings.ToLower(text))
		for _, word := range words {
			sum := sha256.Sum256([]byte(word))
			idx := int(binary.BigEndian.Uint32(sum[:4]) % uint32(dim))
			vector[idx] += 1
		}
		normalize(vector)
		results = append(results, rag.EmbeddingResult{
			Vector:      vector,
			InputTokens: len(words),
			Model:       model,
		})
	}
	return results, nil
}

type VectorStore struct {
	Records []rag.VectorRecord
	Chunks  map[string]rag.Chunk
}

func (v *VectorStore) UpsertChunks(_ context.Context, records []rag.VectorRecord) error {
	v.Records = append(v.Records, records...)
	return nil
}

func (v *VectorStore) Search(_ context.Context, vector []float32, topK int, _ map[string]any) ([]rag.RetrievalHit, error) {
	hits := make([]rag.RetrievalHit, 0, len(v.Records))
	for _, record := range v.Records {
		chunk := rag.Chunk{VectorID: record.ID, Metadata: record.Metadata}
		if v.Chunks != nil {
			if stored, ok := v.Chunks[record.ID]; ok {
				chunk = stored
			}
		}
		hits = append(hits, rag.RetrievalHit{
			Chunk:      chunk,
			DenseScore: cosine(vector, record.Values),
			Source:     "dense",
		})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		return hits[i].DenseScore > hits[j].DenseScore
	})
	if topK > 0 && len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}

func (v *VectorStore) DeleteDocument(_ context.Context, documentID uuid.UUID) error {
	filtered := v.Records[:0]
	for _, record := range v.Records {
		if fmt.Sprint(record.Metadata["document_id"]) != documentID.String() {
			filtered = append(filtered, record)
		}
	}
	v.Records = filtered
	return nil
}

func (v *VectorStore) Healthcheck(context.Context) error {
	return nil
}

type Reranker struct{}

func (Reranker) Rerank(_ context.Context, _ string, documents []rag.RerankDocument, topN int) ([]rag.RerankResult, error) {
	results := make([]rag.RerankResult, 0, len(documents))
	for i, doc := range documents {
		results = append(results, rag.RerankResult{
			ID:    doc.ID,
			Score: 1 / float64(i+1),
			Rank:  i + 1,
		})
	}
	if topN > 0 && len(results) > topN {
		results = results[:topN]
	}
	return results, nil
}

type LLM struct {
	Model string
}

func (l LLM) GenerateAnswer(_ context.Context, req rag.GenerateRequest) (rag.GenerateResponse, error) {
	if len(req.Context) == 0 {
		return rag.GenerateResponse{
			Answer: "I could not find this in the uploaded documents.",
			Model:  l.model(),
		}, nil
	}
	citations := make([]rag.Citation, 0, len(req.Context))
	for _, hit := range req.Context {
		citations = append(citations, rag.Citation{
			ChunkID:     hit.Chunk.ID,
			DocumentID:  hit.Chunk.DocumentID,
			Document:    fmt.Sprint(hit.Chunk.Metadata["filename"]),
			PageNumber:  hit.Chunk.PageNumber,
			Excerpt:     excerpt(hit.Chunk.Content, 240),
			DenseScore:  hit.DenseScore,
			LexicalRank: hit.LexicalRank,
			FusedRank:   hit.FusedRank,
			RerankScore: hit.RerankScore,
			Metadata:    hit.Chunk.Metadata,
		})
	}
	return rag.GenerateResponse{
		Answer:       "Based on the uploaded documents: " + excerpt(req.Context[0].Chunk.Content, 300),
		InputTokens:  len(strings.Fields(req.Query)),
		OutputTokens: len(strings.Fields(req.Context[0].Chunk.Content)),
		Model:        l.model(),
		Citations:    citations,
	}, nil
}

func (l LLM) RewriteQuestion(_ context.Context, question string, history []rag.Message) (string, error) {
	trimmed := strings.TrimSpace(question)
	if len(history) == 0 {
		return trimmed, nil
	}
	return strings.TrimSpace(history[len(history)-1].Content + " " + trimmed), nil
}

func (l LLM) model() string {
	if l.Model == "" {
		return "mock-llm"
	}
	return l.Model
}

func normalize(vector []float32) {
	var sum float64
	for _, value := range vector {
		sum += float64(value * value)
	}
	if sum == 0 {
		return
	}
	norm := float32(math.Sqrt(sum))
	for i := range vector {
		vector[i] /= norm
	}
}

func cosine(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot float64
	for i := range a {
		dot += float64(a[i] * b[i])
	}
	return dot
}

func excerpt(text string, max int) string {
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	return strings.TrimSpace(text[:max]) + "..."
}
