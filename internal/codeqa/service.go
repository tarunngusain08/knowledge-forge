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
	retrievalpkg "github.com/tarunngusain08/knowledge-forge/internal/retrieval"
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
	RerankerEnabled *bool     `json:"reranker_enabled,omitempty"`
}

type AnswerProvenance struct {
	RepositoryID       uuid.UUID      `json:"repository_id"`
	BranchName         string         `json:"branch_name"`
	SnapshotID         uuid.UUID      `json:"snapshot_id"`
	CommitSHA          string         `json:"commit_sha"`
	QueryCategory      string         `json:"query_category"`
	PromptVersion      string         `json:"prompt_version"`
	RetrievalConfig    map[string]any `json:"retrieval_config"`
	RetrievalPath      []string       `json:"retrieval_path"`
	RetrievedChunkIDs  []uuid.UUID    `json:"retrieved_chunk_ids"`
	StageContributions map[string]int `json:"stage_contributions"`
	ContextTokenCount  int            `json:"context_token_count"`
	EstimatedCostUSD   float64        `json:"estimated_cost_usd"`
	Model              string         `json:"model"`
}

type AskResponse struct {
	Answer       string              `json:"answer"`
	Citations    []rag.Citation      `json:"citations"`
	Retrieval    rag.RetrievalResult `json:"retrieval"`
	TraceID      uuid.UUID           `json:"trace_id"`
	Provenance   AnswerProvenance    `json:"provenance"`
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
	policy := retrievalpkg.BuildPolicy(req.Question, retrievalpkg.PolicyOptions{
		TopK:            req.TopK,
		RerankerEnabled: rerankerPreference(req.RerankerEnabled),
	})
	retrieval, err := s.retriever.Retrieve(ctx, rag.RetrievalRequest{
		UserID:             req.UserID,
		RepositoryID:       req.RepositoryID,
		BranchName:         branch,
		Query:              rewritten,
		TopK:               policy.TopK,
		CandidateK:         policy.CandidateK,
		QueryCategory:      policy.Category,
		RetrievalPath:      policy.RetrievalPath,
		RetrievalConfig:    policy.Config,
		ContextTokenBudget: policy.ContextTokenBudget,
		RerankerEnabled:    policy.RerankerEnabled,
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("retrieve repository context: %w", err)
	}
	assembly := rag.AssembleContext(retrieval.RerankedHits, policy.ContextTokenBudget)
	retrieval.ContextTokenCount = assembly.TokenCount
	retrieval.RetrievedChunkIDs = retrievedChunkIDs(assembly.Hits)
	retrieval.StageContributions = withContextContribution(retrieval.StageContributions, len(assembly.Hits))
	generation, err := s.llm.GenerateAnswer(ctx, rag.GenerateRequest{
		Query:          req.Question,
		RewrittenQuery: rewritten,
		Context:        assembly.Hits,
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("generate repository answer: %w", err)
	}
	cost := costs.EstimateUSD(costs.Usage{
		Provider:     "vertex",
		Model:        generation.Model,
		Operation:    "generate",
		InputTokens:  generation.InputTokens,
		OutputTokens: generation.OutputTokens,
	})
	traceID, err := s.repos.CreateRetrievalTrace(ctx, repositories.RetrievalTraceInput{
		UserID:             req.UserID,
		RepositoryID:       req.RepositoryID,
		SnapshotID:         retrieval.SnapshotID,
		BranchName:         branch,
		QueryCategory:      policy.Category,
		RetrievalPath:      policy.RetrievalPath,
		RetrievalConfig:    policy.Config,
		RetrievedChunkIDs:  retrieval.RetrievedChunkIDs,
		StageContributions: retrieval.StageContributions,
		ContextTokenCount:  assembly.TokenCount,
		PromptVersion:      rag.PromptVersion,
		GenerationModel:    generation.Model,
		EstimatedCostUSD:   cost,
		OriginalQuery:      req.Question,
		RewrittenQuery:     rewritten,
		TopK:               policy.TopK,
		RerankerEnabled:    policy.RerankerEnabled,
		DenseHits:          retrieval.DenseHits,
		LexicalHits:        retrieval.LexicalHits,
		SymbolHits:         retrieval.SymbolHits,
		GraphHits:          retrieval.GraphHits,
		FusedHits:          retrieval.FusedHits,
		RerankedHits:       retrieval.RerankedHits,
		PromptPreview:      rag.BuildGroundedPrompt(rag.GenerateRequest{RewrittenQuery: rewritten, Context: assembly.Hits}),
		LatencyMS:          retrieval.Latency.Milliseconds(),
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("create retrieval trace: %w", err)
	}
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
		Answer:    generation.Answer,
		Citations: generation.Citations,
		Retrieval: retrieval,
		TraceID:   traceID,
		Provenance: AnswerProvenance{
			RepositoryID:       req.RepositoryID,
			BranchName:         branch,
			SnapshotID:         retrieval.SnapshotID,
			CommitSHA:          retrieval.CommitSHA,
			QueryCategory:      policy.Category,
			PromptVersion:      rag.PromptVersion,
			RetrievalConfig:    policy.Config,
			RetrievalPath:      policy.RetrievalPath,
			RetrievedChunkIDs:  retrieval.RetrievedChunkIDs,
			StageContributions: retrieval.StageContributions,
			ContextTokenCount:  assembly.TokenCount,
			EstimatedCostUSD:   cost,
			Model:              generation.Model,
		},
		Model:        generation.Model,
		InputTokens:  generation.InputTokens,
		OutputTokens: generation.OutputTokens,
	}, nil
}

func nullableUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func rerankerPreference(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}

func retrievedChunkIDs(hits []rag.RetrievalHit) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(hits))
	for _, hit := range hits {
		if hit.Chunk.ID != uuid.Nil {
			ids = append(ids, hit.Chunk.ID)
		}
	}
	return ids
}

func withContextContribution(stages map[string]int, count int) map[string]int {
	if stages == nil {
		stages = map[string]int{}
	}
	stages["context"] = count
	return stages
}
