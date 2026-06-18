package codeqa

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	assembly := assembleContextWithRequiredEvidence(req.Question, retrieval.RerankedHits, policy.ContextTokenBudget)
	retrieval.ContextTokenCount = assembly.TokenCount
	retrieval.RetrievedChunkIDs = retrievedChunkIDs(assembly.Hits)
	retrieval.StageContributions = withContextContribution(retrieval.StageContributions, len(assembly.Hits))
	support := evaluateAnswerSupport(req.Question, policy.Category, assembly.Hits)
	retrieval, assembly, support, err = s.completeMissingSupportEvidence(ctx, req, branch, policy, retrieval, assembly, support)
	if err != nil {
		return AskResponse{}, err
	}
	policy.Config = retrievalConfigWithSupportGate(policy.Config, support)
	retrieval.RetrievalConfig = policy.Config

	generation := rag.GenerateResponse{
		Answer:    refusalAnswer,
		Model:     "support-gate",
		Citations: []rag.Citation{},
	}
	var cost float64
	if support.Answerable {
		generation, err = s.llm.GenerateAnswer(ctx, rag.GenerateRequest{
			Query:          req.Question,
			RewrittenQuery: rewritten,
			Context:        assembly.Hits,
		})
		if err != nil {
			return AskResponse{}, fmt.Errorf("generate repository answer: %w", err)
		}
		if generation.Citations == nil {
			generation.Citations = []rag.Citation{}
		}
		cost = costs.EstimateUSD(costs.Usage{
			Provider:     "vertex",
			Model:        generation.Model,
			Operation:    "generate",
			InputTokens:  generation.InputTokens,
			OutputTokens: generation.OutputTokens,
		})
	}
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
		LatencyMS:          retrieval.LatencyMilliseconds(),
	})
	if err != nil {
		return AskResponse{}, fmt.Errorf("create retrieval trace: %w", err)
	}
	if s.costs != nil && support.Answerable {
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

func (s *Service) completeMissingSupportEvidence(ctx context.Context, req AskRequest, branch string, policy retrievalpkg.Policy, retrieval rag.RetrievalResult, assembly rag.ContextAssembly, support supportGateResult) (rag.RetrievalResult, rag.ContextAssembly, supportGateResult, error) {
	groups := recoverableEvidenceGroups(support.MissingEvidence)
	if support.Answerable || len(groups) == 0 {
		return retrieval, assembly, support, nil
	}
	followUps := 0
	for _, group := range groups {
		query := evidenceGroupFollowUpQuery(group)
		if query == "" {
			continue
		}
		followUp, err := s.retriever.Retrieve(ctx, rag.RetrievalRequest{
			UserID:             req.UserID,
			RepositoryID:       req.RepositoryID,
			BranchName:         branch,
			Query:              query,
			TopK:               3,
			CandidateK:         8,
			QueryCategory:      policy.Category,
			RetrievalPath:      append([]string{}, append(policy.RetrievalPath, "evidence_followup")...),
			RetrievalConfig:    policy.Config,
			ContextTokenBudget: policy.ContextTokenBudget,
			RerankerEnabled:    policy.RerankerEnabled,
		})
		if err != nil {
			return retrieval, assembly, support, fmt.Errorf("retrieve missing evidence group %s: %w", group, err)
		}
		retrieval = mergeRetrievalResults(retrieval, followUp)
		assembly = assembleContextWithRequiredEvidence(req.Question, retrieval.RerankedHits, policy.ContextTokenBudget)
		retrieval.ContextTokenCount = assembly.TokenCount
		retrieval.RetrievedChunkIDs = retrievedChunkIDs(assembly.Hits)
		retrieval.StageContributions = withContextContribution(retrieval.StageContributions, len(assembly.Hits))
		retrieval.StageContributions["evidence_followup"] += len(followUp.RerankedHits)
		support = evaluateAnswerSupport(req.Question, policy.Category, assembly.Hits)
		followUps++
		if support.Answerable || followUps >= 4 {
			break
		}
	}
	return retrieval, assembly, support, nil
}

func evidenceGroupFollowUpQuery(group string) string {
	switch group {
	case "api_router":
		return "HTTP API router route registration"
	case "chat_handler":
		return "chat session HTTP handler"
	case "auth_service":
		return "authentication service JWT implementation"
	case "database_connection":
		return "database connection setup pgx sqlc generated db"
	case "report_generator":
		return "deep-dive report generation sections citations"
	case "repo_qa_service":
		return "repository QA service report orchestration"
	case "evidence_quality":
		return "report evidence quality citations coverage"
	case "retrieval_source", "dense_or_code_retrieval":
		return "repository retrieval candidate generation source"
	case "rag_context", "context_assembly":
		return "RAG context assembly answer generation"
	case "lexical_retrieval", "postgres_fts":
		return "PostgreSQL FTS lexical retrieval tsquery"
	case "repository_api":
		return "repository registration API handler"
	case "repository_store":
		return "repository registration store metadata"
	case "fusion":
		return "RRF fusion ranked candidates"
	default:
		return ""
	}
}

func mergeRetrievalResults(base, extra rag.RetrievalResult) rag.RetrievalResult {
	base.DenseHits = mergeHits(base.DenseHits, extra.DenseHits)
	base.LexicalHits = mergeHits(base.LexicalHits, extra.LexicalHits)
	base.SymbolHits = mergeHits(base.SymbolHits, extra.SymbolHits)
	base.GraphHits = mergeHits(base.GraphHits, extra.GraphHits)
	base.FusedHits = mergeHits(base.FusedHits, extra.FusedHits)
	base.RerankedHits = mergeHits(base.RerankedHits, extra.RerankedHits)
	if base.StageContributions == nil {
		base.StageContributions = map[string]int{}
	}
	for stage, count := range extra.StageContributions {
		base.StageContributions[stage] += count
	}
	if base.LatencyMS == 0 {
		base.LatencyMS = extra.LatencyMS
	} else {
		base.LatencyMS += extra.LatencyMS
	}
	return base
}

func mergeHits(primary, secondary []rag.RetrievalHit) []rag.RetrievalHit {
	seen := map[string]bool{}
	merged := make([]rag.RetrievalHit, 0, len(primary)+len(secondary))
	for _, hit := range append(primary, secondary...) {
		key := retrievalHitKey(hit)
		if seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, hit)
	}
	return merged
}

func retrievalHitKey(hit rag.RetrievalHit) string {
	if hit.Chunk.ID != uuid.Nil {
		return hit.Chunk.ID.String()
	}
	path := strings.TrimSpace(fmt.Sprint(hit.Chunk.Metadata["path"]))
	start := fmt.Sprint(hit.Chunk.Metadata["start_line"])
	end := fmt.Sprint(hit.Chunk.Metadata["end_line"])
	return path + ":" + start + ":" + end + ":" + hit.Chunk.Content
}

func assembleContextWithRequiredEvidence(question string, hits []rag.RetrievalHit, maxTokens int) rag.ContextAssembly {
	assembly := rag.AssembleContext(hits, maxTokens)
	required := requiredEvidenceGroups(question)
	if len(required) == 0 {
		return assembly
	}
	_, missing := evidenceGroupSupport(required, assembly.Hits)
	if len(missing) == 0 {
		return assembly
	}
	_, unavailable := evidenceGroupSupport(missing, hits)
	if len(unavailable) > 0 {
		return assembly
	}
	candidate := rag.AssembleContext(prioritizeRequiredEvidenceHits(required, hits, maxTokens), maxTokens)
	_, candidateMissing := evidenceGroupSupport(required, candidate.Hits)
	if len(candidateMissing) < len(missing) {
		return candidate
	}
	return assembly
}

func prioritizeRequiredEvidenceHits(required []string, hits []rag.RetrievalHit, maxTokens int) []rag.RetrievalHit {
	if len(required) == 0 || len(hits) == 0 {
		return hits
	}
	selected := make([]rag.RetrievalHit, 0, len(required))
	seen := map[string]bool{}
	for _, group := range required {
		for _, hit := range hits {
			if !hitSupportsEvidenceGroup(hit, group) {
				continue
			}
			key := retrievalHitKey(hit)
			if seen[key] {
				break
			}
			seen[key] = true
			selected = append(selected, hit)
			break
		}
	}
	if len(selected) == 0 {
		return hits
	}
	if maxTokens <= 0 {
		maxTokens = 2400
	}
	budgetPerHit := maxTokens / len(selected)
	if budgetPerHit <= 0 {
		budgetPerHit = maxTokens
	}
	for idx := range selected {
		selected[idx] = compactHitForTokenBudget(selected[idx], budgetPerHit)
	}
	return mergeHits(selected, hits)
}

func hitSupportsEvidenceGroup(hit rag.RetrievalHit, group string) bool {
	for available := range evidenceGroupSet([]rag.RetrievalHit{hit}) {
		if available == group {
			return true
		}
	}
	return false
}

func compactHitForTokenBudget(hit rag.RetrievalHit, tokenBudget int) rag.RetrievalHit {
	if tokenBudget <= 0 {
		return hit
	}
	current := hit.Chunk.TokenCount
	if current <= 0 {
		current = estimateContextTokens(hit.Chunk.Content)
	}
	if current <= tokenBudget {
		hit.Chunk.TokenCount = current
		return hit
	}
	words := strings.Fields(hit.Chunk.Content)
	if len(words) == 0 {
		hit.Chunk.TokenCount = 0
		return hit
	}
	maxWords := (tokenBudget - 1) * 3 / 4
	if maxWords < 1 {
		maxWords = 1
	}
	if len(words) > maxWords {
		hit.Chunk.Content = strings.Join(words[:maxWords], " ")
	}
	hit.Chunk.TokenCount = estimateContextTokens(hit.Chunk.Content)
	return hit
}

func estimateContextTokens(text string) int {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return 0
	}
	return (len(fields) * 4 / 3) + 1
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
