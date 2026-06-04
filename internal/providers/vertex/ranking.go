package vertex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2/google"

	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type Ranker struct {
	ProjectID string
	Location  string
	Model     string
	Client    *http.Client
}

func NewRanker(ctx context.Context, projectID, location, model string) (*Ranker, error) {
	if projectID == "" {
		return nil, fmt.Errorf("google cloud project is required")
	}
	if location == "" {
		location = "global"
	}
	if model == "" {
		model = "semantic-ranker-default@latest"
	}
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("create google auth client: %w", err)
	}
	return &Ranker{ProjectID: projectID, Location: location, Model: model, Client: client}, nil
}

func (r *Ranker) Rerank(ctx context.Context, query string, documents []rag.RerankDocument, topN int) ([]rag.RerankResult, error) {
	records := make([]map[string]string, 0, len(documents))
	for _, doc := range documents {
		records = append(records, map[string]string{"id": doc.ID, "content": doc.Content})
	}
	body, err := json.Marshal(map[string]any{
		"model":   r.Model,
		"query":   query,
		"records": records,
		"topN":    topN,
	})
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf(
		"https://discoveryengine.googleapis.com/v1/projects/%s/locations/%s/rankingConfigs/default_ranking_config:rank",
		r.ProjectID,
		r.Location,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("vertex ranking request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vertex ranking status %d", resp.StatusCode)
	}
	var decoded rankResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode ranking response: %w", err)
	}
	results := make([]rag.RerankResult, 0, len(decoded.Records))
	for i, record := range decoded.Records {
		results = append(results, rag.RerankResult{ID: record.ID, Score: record.Score, Rank: i + 1})
	}
	return results, nil
}

func (r *Ranker) httpClient() *http.Client {
	if r.Client != nil {
		return r.Client
	}
	return http.DefaultClient
}

type rankResponse struct {
	Records []struct {
		ID    string  `json:"id"`
		Score float64 `json:"score"`
	} `json:"records"`
}
