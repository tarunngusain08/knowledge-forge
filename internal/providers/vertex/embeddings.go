package vertex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2/google"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

type Embeddings struct {
	ProjectID string
	Location  string
	Model     string
	Client    *http.Client
}

func NewEmbeddings(ctx context.Context, projectID, location, model string) (*Embeddings, error) {
	if projectID == "" {
		return nil, fmt.Errorf("google cloud project is required")
	}
	if location == "" {
		location = "us-central1"
	}
	if model == "" {
		model = "gemini-embedding-001"
	}
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("create google auth client: %w", err)
	}
	return &Embeddings{ProjectID: projectID, Location: location, Model: model, Client: client}, nil
}

func (e *Embeddings) EmbedDocuments(ctx context.Context, texts []string) ([]rag.EmbeddingResult, error) {
	return e.embed(ctx, texts, rag.TaskRetrievalDocument)
}

func (e *Embeddings) EmbedQuery(ctx context.Context, text string) (rag.EmbeddingResult, error) {
	results, err := e.embed(ctx, []string{text}, rag.TaskRetrievalQuery)
	if err != nil {
		return rag.EmbeddingResult{}, err
	}
	return results[0], nil
}

func (e *Embeddings) embed(ctx context.Context, texts []string, taskType string) ([]rag.EmbeddingResult, error) {
	instances := make([]map[string]string, 0, len(texts))
	for _, text := range texts {
		instances = append(instances, map[string]string{
			"content":   text,
			"task_type": taskType,
		})
	}
	body, err := json.Marshal(map[string]any{"instances": instances})
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		e.Location,
		e.ProjectID,
		e.Location,
		e.Model,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("vertex embedding request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vertex embedding status %d", resp.StatusCode)
	}

	var decoded embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode vertex embedding response: %w", err)
	}
	results := make([]rag.EmbeddingResult, 0, len(decoded.Predictions))
	for _, prediction := range decoded.Predictions {
		results = append(results, rag.EmbeddingResult{
			Vector:      prediction.Embeddings.Values,
			InputTokens: prediction.Embeddings.Statistics.TokenCount,
			Model:       e.Model,
		})
	}
	if len(results) != len(texts) {
		return nil, fmt.Errorf("vertex returned %d embeddings for %d inputs", len(results), len(texts))
	}
	return results, nil
}

func (e *Embeddings) httpClient() *http.Client {
	if e.Client != nil {
		return e.Client
	}
	return http.DefaultClient
}

type embeddingResponse struct {
	Predictions []struct {
		Embeddings struct {
			Values     []float32 `json:"values"`
			Statistics struct {
				TokenCount int `json:"token_count"`
			} `json:"statistics"`
		} `json:"embeddings"`
	} `json:"predictions"`
}

func CleanModelName(model string) string {
	return strings.TrimSpace(model)
}
