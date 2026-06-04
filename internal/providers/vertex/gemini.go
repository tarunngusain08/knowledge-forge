package vertex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2/google"

	"github.com/tarunngusain08/RAG-bot/internal/rag"
)

type Gemini struct {
	ProjectID string
	Location  string
	Model     string
	Client    *http.Client
}

func NewGemini(ctx context.Context, projectID, location, model string) (*Gemini, error) {
	if projectID == "" {
		return nil, fmt.Errorf("google cloud project is required")
	}
	if location == "" {
		location = "us-central1"
	}
	if model == "" {
		model = "gemini-2.5-flash"
	}
	client, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("create google auth client: %w", err)
	}
	return &Gemini{ProjectID: projectID, Location: location, Model: model, Client: client}, nil
}

func (g *Gemini) GenerateAnswer(ctx context.Context, req rag.GenerateRequest) (rag.GenerateResponse, error) {
	prompt := rag.BuildGroundedPrompt(req)
	text, usage, err := g.generate(ctx, prompt)
	if err != nil {
		return rag.GenerateResponse{}, err
	}
	return rag.GenerateResponse{
		Answer:       strings.TrimSpace(text),
		InputTokens:  usage.PromptTokenCount,
		OutputTokens: usage.CandidatesTokenCount,
		Model:        g.Model,
		Citations:    citationsFromContext(req.Context),
	}, nil
}

func (g *Gemini) RewriteQuestion(ctx context.Context, question string, history []rag.Message) (string, error) {
	if len(history) == 0 {
		return strings.TrimSpace(question), nil
	}
	var b strings.Builder
	b.WriteString("Rewrite the latest question as a standalone search query. Do not answer it.\n\nConversation:\n")
	for _, msg := range history {
		b.WriteString(msg.Role)
		b.WriteString(": ")
		b.WriteString(msg.Content)
		b.WriteString("\n")
	}
	b.WriteString("Latest question: ")
	b.WriteString(question)
	text, _, err := g.generate(ctx, b.String())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func (g *Gemini) generate(ctx context.Context, prompt string) (string, usageMetadata, error) {
	body, err := json.Marshal(map[string]any{
		"contents": []map[string]any{{
			"role": "user",
			"parts": []map[string]string{{
				"text": prompt,
			}},
		}},
		"generationConfig": map[string]any{
			"temperature":     0.2,
			"maxOutputTokens": 1024,
		},
	})
	if err != nil {
		return "", usageMetadata{}, err
	}
	endpoint := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		g.Location,
		g.ProjectID,
		g.Location,
		g.Model,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", usageMetadata{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.httpClient().Do(req)
	if err != nil {
		return "", usageMetadata{}, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", usageMetadata{}, fmt.Errorf("gemini status %d", resp.StatusCode)
	}
	var decoded generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", usageMetadata{}, fmt.Errorf("decode gemini response: %w", err)
	}
	if len(decoded.Candidates) == 0 || len(decoded.Candidates[0].Content.Parts) == 0 {
		return "", decoded.UsageMetadata, nil
	}
	return decoded.Candidates[0].Content.Parts[0].Text, decoded.UsageMetadata, nil
}

func (g *Gemini) httpClient() *http.Client {
	if g.Client != nil {
		return g.Client
	}
	return http.DefaultClient
}

func citationsFromContext(hits []rag.RetrievalHit) []rag.Citation {
	citations := make([]rag.Citation, 0, len(hits))
	for _, hit := range hits {
		citations = append(citations, rag.Citation{
			ChunkID:     hit.Chunk.ID,
			DocumentID:  hit.Chunk.DocumentID,
			Document:    fmt.Sprint(hit.Chunk.Metadata["filename"]),
			PageNumber:  hit.Chunk.PageNumber,
			Excerpt:     excerpt(hit.Chunk.Content, 320),
			DenseScore:  hit.DenseScore,
			LexicalRank: hit.LexicalRank,
			FusedRank:   hit.FusedRank,
			RerankScore: hit.RerankScore,
			Metadata:    hit.Chunk.Metadata,
		})
	}
	return citations
}

func excerpt(text string, max int) string {
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	return strings.TrimSpace(text[:max]) + "..."
}

type generateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata usageMetadata `json:"usageMetadata"`
}

type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}
