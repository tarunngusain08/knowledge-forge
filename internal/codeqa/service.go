package codeqa

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/costs"
	"github.com/tarunngusain08/knowledge-forge/internal/db"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
)

type RepositoryStore interface {
	GetRepositoryForUser(ctx context.Context, ownerUserID, id uuid.UUID) (codeintel.Repository, error)
	CreateRetrievalTrace(ctx context.Context, input repositories.RetrievalTraceInput) (uuid.UUID, error)
}

type CostStore interface {
	CreateTokenCostEvent(ctx context.Context, arg db.CreateTokenCostEventParams) (db.TokenCostEvent, error)
}

type Service struct {
	repos     RepositoryStore
	costs     CostStore
	llm       rag.LLMProvider
	retriever rag.Retriever
}

type AskRequest struct {
	UserID          uuid.UUID `json:"user_id"`
	RepositoryID    uuid.UUID `json:"repository_id"`
	BranchName      string    `json:"branch_name"`
	Question        string    `json:"question"`
	TopK            int       `json:"top_k"`
	RerankerEnabled bool      `json:"reranker_enabled"`
}

type AskResponse struct {
	Answer       string              `json:"answer"`
	Citations    []rag.Citation      `json:"citations"`
	Retrieval    rag.RetrievalResult `json:"retrieval"`
	TraceID      uuid.UUID           `json:"trace_id"`
	Model        string              `json:"model"`
	InputTokens  int                 `json:"input_tokens"`
	OutputTokens int                 `json:"output_tokens"`
}

func NewService(repos RepositoryStore, costStore CostStore, llm rag.LLMProvider, retriever rag.Retriever) *Service {
	return &Service{repos: repos, costs: costStore, llm: llm, retriever: retriever}
}

func (s *Service) Ask(ctx context.Context, req AskRequest) (AskResponse, error) {
	repo, err := s.repos.GetRepositoryForUser(ctx, req.UserID, req.RepositoryID)
	if err != nil {
		return AskResponse{}, fmt.Errorf("get repository: %w", err)
	}
	branch := req.BranchName
	if branch == "" {
		branch = repo.DefaultBranch
	}
	rewritten, err := s.llm.RewriteQuestion(ctx, req.Question, nil)
	if err != nil {
		return AskResponse{}, fmt.Errorf("rewrite question: %w", err)
	}
	retrieval, err := s.retriever.Retrieve(ctx, rag.RetrievalRequest{
		UserID:          req.UserID,
		RepositoryID:    req.RepositoryID,
		BranchName:      branch,
		Query:           rewritten,
		TopK:            req.TopK,
		RerankerEnabled: req.RerankerEnabled,
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("retrieve repository context: %w", err)
	}
	generation, err := s.llm.GenerateAnswer(ctx, rag.GenerateRequest{
		Query:          req.Question,
		RewrittenQuery: rewritten,
		Context:        retrieval.RerankedHits,
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("generate repository answer: %w", err)
	}
	traceID, err := s.repos.CreateRetrievalTrace(ctx, repositories.RetrievalTraceInput{
		UserID:          req.UserID,
		RepositoryID:    req.RepositoryID,
		SnapshotID:      retrieval.SnapshotID,
		BranchName:      branch,
		OriginalQuery:   req.Question,
		RewrittenQuery:  rewritten,
		TopK:            effectiveTopK(req.TopK),
		RerankerEnabled: req.RerankerEnabled,
		DenseHits:       retrieval.DenseHits,
		LexicalHits:     retrieval.LexicalHits,
		SymbolHits:      retrieval.SymbolHits,
		GraphHits:       retrieval.GraphHits,
		FusedHits:       retrieval.FusedHits,
		RerankedHits:    retrieval.RerankedHits,
		PromptPreview:   rag.BuildGroundedPrompt(rag.GenerateRequest{RewrittenQuery: rewritten, Context: retrieval.RerankedHits}),
		LatencyMS:       retrieval.Latency.Milliseconds(),
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("create retrieval trace: %w", err)
	}
	cost := costs.EstimateUSD(costs.Usage{
		Provider:     "vertex",
		Model:        generation.Model,
		Operation:    "generate",
		InputTokens:  generation.InputTokens,
		OutputTokens: generation.OutputTokens,
	})
	if s.costs != nil {
		_, _ = s.costs.CreateTokenCostEvent(ctx, db.CreateTokenCostEventParams{
			UserID:           nullableUUID(req.UserID),
			Provider:         "vertex",
			Model:            generation.Model,
			Operation:        "repo_generate",
			InputTokens:      int32(generation.InputTokens),
			OutputTokens:     int32(generation.OutputTokens),
			EstimatedCostUsd: strconv.FormatFloat(cost, 'f', 6, 64),
		})
	}
	return AskResponse{
		Answer:       generation.Answer,
		Citations:    generation.Citations,
		Retrieval:    retrieval,
		TraceID:      traceID,
		Model:        generation.Model,
		InputTokens:  generation.InputTokens,
		OutputTokens: generation.OutputTokens,
	}, nil
}

func effectiveTopK(topK int) int {
	if topK <= 0 {
		return 8
	}
	return topK
}

func nullableUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}
