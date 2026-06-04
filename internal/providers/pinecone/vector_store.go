package pinecone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type VectorStore struct {
	Host      string
	APIKey    string
	Namespace string
	Client    *http.Client
}

func (p *VectorStore) UpsertChunks(ctx context.Context, records []rag.VectorRecord) error {
	if len(records) == 0 {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"namespace": p.Namespace,
		"vectors":   records,
	})
	if err != nil {
		return err
	}
	return p.post(ctx, "/vectors/upsert", body, nil)
}

func (p *VectorStore) Search(ctx context.Context, vector []float32, topK int, filter map[string]any) ([]rag.RetrievalHit, error) {
	body, err := json.Marshal(map[string]any{
		"namespace":       p.Namespace,
		"vector":          vector,
		"topK":            topK,
		"includeMetadata": true,
		"filter":          filter,
	})
	if err != nil {
		return nil, err
	}
	var decoded queryResponse
	if err := p.post(ctx, "/query", body, &decoded); err != nil {
		return nil, err
	}
	hits := make([]rag.RetrievalHit, 0, len(decoded.Matches))
	for _, match := range decoded.Matches {
		hits = append(hits, rag.RetrievalHit{
			Chunk: rag.Chunk{
				VectorID: match.ID,
				Metadata: match.Metadata,
			},
			DenseScore: match.Score,
			Source:     "dense",
		})
	}
	return hits, nil
}

func (p *VectorStore) DeleteDocument(ctx context.Context, documentID uuid.UUID) error {
	body, err := json.Marshal(map[string]any{
		"namespace": p.Namespace,
		"filter": map[string]any{
			"document_id": map[string]string{"$eq": documentID.String()},
		},
	})
	if err != nil {
		return err
	}
	return p.post(ctx, "/vectors/delete", body, nil)
}

func (p *VectorStore) Healthcheck(ctx context.Context) error {
	return p.post(ctx, "/describe_index_stats", []byte(`{}`), nil)
}

func (p *VectorStore) post(ctx context.Context, path string, body []byte, out any) error {
	if p.Host == "" || p.APIKey == "" {
		return fmt.Errorf("pinecone host and api key are required")
	}
	url := strings.TrimRight(p.Host, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", p.APIKey)
	req.Header.Set("X-Pinecone-API-Version", "2025-01")
	resp, err := p.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("pinecone request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("pinecone status %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode pinecone response: %w", err)
		}
	}
	return nil
}

func (p *VectorStore) httpClient() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

type queryResponse struct {
	Matches []struct {
		ID       string         `json:"id"`
		Score    float64        `json:"score"`
		Metadata map[string]any `json:"metadata"`
	} `json:"matches"`
}
