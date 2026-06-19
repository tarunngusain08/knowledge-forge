package codeqa

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
	"github.com/tarunngusain08/knowledge-forge/internal/repositories"
)

func TestAskRefusesUnsupportedQuestionBeforeGeneration(t *testing.T) {
	userID := uuid.New()
	repoID := uuid.New()
	snapshotID := uuid.New()
	store := &codeQATestStore{
		repo: codeintel.Repository{
			ID:            repoID,
			OwnerUserID:   userID,
			DefaultBranch: "main",
		},
	}
	llm := &codeQATestLLM{}
	service := NewService(store, nil, llm, codeQATestRetriever{
		repositoryID: repoID,
		snapshotID:   snapshotID,
		hits: []rag.RetrievalHit{
			codeQATestHit(repoID, snapshotID, "internal/auth/service.go", "package auth\nfunc Login() error { return nil }"),
		},
	})

	response, err := service.Ask(context.Background(), AskRequest{
		UserID:       userID,
		RepositoryID: repoID,
		Question:     "Which service processes payroll?",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if response.Answer != refusalAnswer {
		t.Fatalf("answer = %q", response.Answer)
	}
	if llm.generated != 0 {
		t.Fatalf("generation should be skipped for unsupported evidence")
	}
	if len(response.Retrieval.RerankedHits) != 0 || len(response.Retrieval.DenseHits) != 0 {
		t.Fatalf("refused response exposed retrieval hits: %#v", response.Retrieval)
	}
	if len(response.Provenance.RetrievedChunkIDs) != 0 || response.Provenance.ContextTokenCount != 0 {
		t.Fatalf("refused response provenance exposed retrieval context: %#v", response.Provenance)
	}
	if store.trace.PromptPreview != "" {
		t.Fatalf("refused trace exposed prompt preview: %q", store.trace.PromptPreview)
	}
	if traceHits, ok := store.trace.RerankedHits.([]rag.RetrievalHit); ok && len(traceHits) != 0 {
		t.Fatalf("refused trace exposed reranked hits: %#v", traceHits)
	}
	if len(store.trace.RetrievedChunkIDs) != 0 || store.trace.ContextTokenCount != 0 {
		t.Fatalf("refused trace exposed retrieval context: %#v", store.trace)
	}
	gate := supportGateFromTrace(t, store.trace)
	if gate.Answerable {
		t.Fatalf("support gate = %#v", gate)
	}
	if !contains(gate.MissingTerms, "payroll") {
		t.Fatalf("missing terms = %#v", gate.MissingTerms)
	}
}

func TestAskKeepsAnswerableRepositoryQuestions(t *testing.T) {
	userID := uuid.New()
	repoID := uuid.New()
	snapshotID := uuid.New()
	store := &codeQATestStore{
		repo: codeintel.Repository{
			ID:            repoID,
			OwnerUserID:   userID,
			DefaultBranch: "main",
		},
	}
	llm := &codeQATestLLM{}
	service := NewService(store, nil, llm, codeQATestRetriever{
		repositoryID: repoID,
		snapshotID:   snapshotID,
		hits: []rag.RetrievalHit{
			codeQATestHit(repoID, snapshotID, "internal/auth/service.go", "package auth\ntype AuthService struct{}\nfunc Login() error { return nil }"),
		},
	})

	response, err := service.Ask(context.Background(), AskRequest{
		UserID:       userID,
		RepositoryID: repoID,
		Question:     "Where is authentication implemented?",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if response.Answer == refusalAnswer {
		t.Fatalf("known answer was refused")
	}
	if llm.generated != 1 {
		t.Fatalf("generation count = %d", llm.generated)
	}
	gate := supportGateFromTrace(t, store.trace)
	if !gate.Answerable {
		t.Fatalf("support gate = %#v", gate)
	}
	if !contains(gate.MatchedTerms, "auth") {
		t.Fatalf("matched terms = %#v", gate.MatchedTerms)
	}
}

func TestSupportGateRejectsDeletedSymbolLookup(t *testing.T) {
	gate := evaluateAnswerSupport("Where is LegacyAuthManager implemented?", "exact_symbol_file_lookup", []rag.RetrievalHit{
		codeQATestHit(uuid.New(), uuid.New(), "internal/auth/service.go", "package auth\ntype AuthService struct{}"),
	})

	if gate.Answerable {
		t.Fatalf("deleted symbol lookup should not be answerable: %#v", gate)
	}
	if gate.Reason != "missing_identifier" {
		t.Fatalf("reason = %s", gate.Reason)
	}
}

func TestSupportGateRefusesBusinessDomainFromPathShortcut(t *testing.T) {
	gate := evaluateAnswerSupport("What production revenue API exposes?", "implementation_question", []rag.RetrievalHit{
		codeQATestHit(uuid.New(), uuid.New(), "cmd/api/main.go", "package main\n// API server bootstrap. No revenue endpoint is defined here."),
	})

	if gate.Answerable {
		t.Fatalf("revenue API shortcut should not be answerable: %#v", gate)
	}
	if gate.Reason != "missing_domain_terms" {
		t.Fatalf("reason = %s", gate.Reason)
	}
	if !contains(gate.MatchedEvidence, "api_path_match") {
		t.Fatalf("matched evidence should preserve shortcut diagnostic: %#v", gate.MatchedEvidence)
	}
	if !contains(gate.MissingEvidence, "revenue_domain_evidence") {
		t.Fatalf("missing evidence = %#v", gate.MissingEvidence)
	}
}

func TestSupportGateAllowsCompleteEvidenceGroupsWithoutTermOverlap(t *testing.T) {
	repoID := uuid.New()
	snapshotID := uuid.New()
	gate := evaluateAnswerSupport("How does repository registration work?", "architecture_explanation", []rag.RetrievalHit{
		codeQATestHit(repoID, snapshotID, "internal/httpapi/repository_handlers.go", "package httpapi\nfunc CreateRepository() {}"),
		codeQATestHit(repoID, snapshotID, "internal/repositories/service.go", "package repositories\nfunc Create() {}"),
	})

	if !gate.Answerable {
		t.Fatalf("complete repository registration evidence should be answerable: %#v", gate)
	}
	if gate.Reason != "repository_supported_fact" {
		t.Fatalf("reason = %s", gate.Reason)
	}
	if len(gate.MissingEvidence) != 0 {
		t.Fatalf("missing evidence = %#v", gate.MissingEvidence)
	}
	if !contains(gate.MatchedEvidence, "repository_api") || !contains(gate.MatchedEvidence, "repository_store") {
		t.Fatalf("matched evidence = %#v", gate.MatchedEvidence)
	}
}

func TestAskCompletesMissingEvidenceGroupsBeforeGeneration(t *testing.T) {
	userID := uuid.New()
	repoID := uuid.New()
	snapshotID := uuid.New()
	store := &codeQATestStore{
		repo: codeintel.Repository{
			ID:            repoID,
			OwnerUserID:   userID,
			DefaultBranch: "main",
		},
	}
	llm := &codeQATestLLM{}
	service := NewService(store, nil, llm, codeQAFollowUpRetriever{
		repositoryID: repoID,
		snapshotID:   snapshotID,
	})

	response, err := service.Ask(context.Background(), AskRequest{
		UserID:       userID,
		RepositoryID: repoID,
		Question:     "Explain authentication and database wiring.",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if response.Answer == refusalAnswer {
		t.Fatalf("multi-concept answer was refused")
	}
	files := citationFiles(response.Citations)
	if !contains(files, "internal/auth/auth.go") || !contains(files, "internal/database/connect.go") {
		t.Fatalf("citations did not include auth and database evidence: %#v", files)
	}
	gate := supportGateFromTrace(t, store.trace)
	if !gate.Answerable {
		t.Fatalf("support gate = %#v", gate)
	}
	if !contains(gate.MatchedEvidence, "auth_service") || !contains(gate.MatchedEvidence, "database_connection") {
		t.Fatalf("matched evidence = %#v", gate.MatchedEvidence)
	}
}

func TestAskRetainsReportOrchestrationEvidenceAfterFollowUp(t *testing.T) {
	userID := uuid.New()
	repoID := uuid.New()
	snapshotID := uuid.New()
	store := &codeQATestStore{
		repo: codeintel.Repository{
			ID:            repoID,
			OwnerUserID:   userID,
			DefaultBranch: "main",
		},
	}
	llm := &codeQATestLLM{}
	service := NewService(store, nil, llm, codeQAReportFollowUpRetriever{
		repositoryID: repoID,
		snapshotID:   snapshotID,
	})

	response, err := service.Ask(context.Background(), AskRequest{
		UserID:       userID,
		RepositoryID: repoID,
		Question:     "How are deep-dive reports generated?",
	})
	if err != nil {
		t.Fatalf("Ask returned error: %v", err)
	}
	if response.Answer == refusalAnswer {
		t.Fatalf("report generation answer was refused")
	}
	files := citationFiles(response.Citations)
	if !contains(files, "internal/codeqa/reports.go") || !contains(files, "internal/codeqa/service.go") {
		t.Fatalf("citations did not include report and orchestration evidence: %#v", files)
	}
	gate := supportGateFromTrace(t, store.trace)
	if !gate.Answerable {
		t.Fatalf("support gate = %#v", gate)
	}
	if !contains(gate.MatchedEvidence, "report_generator") || !contains(gate.MatchedEvidence, "repo_qa_service") {
		t.Fatalf("matched evidence = %#v", gate.MatchedEvidence)
	}
}

type codeQATestStore struct {
	repo  codeintel.Repository
	trace repositories.RetrievalTraceInput
}

func (s *codeQATestStore) GetRepositoryForUser(_ context.Context, _, _ uuid.UUID) (codeintel.Repository, error) {
	return s.repo, nil
}

func (s *codeQATestStore) CreateRetrievalTrace(_ context.Context, input repositories.RetrievalTraceInput) (uuid.UUID, error) {
	s.trace = input
	return uuid.New(), nil
}

type codeQATestLLM struct {
	generated int
}

func (l *codeQATestLLM) RewriteQuestion(_ context.Context, question string, _ []rag.Message) (string, error) {
	return question, nil
}

func (l *codeQATestLLM) GenerateAnswer(_ context.Context, req rag.GenerateRequest) (rag.GenerateResponse, error) {
	l.generated++
	citations := make([]rag.Citation, 0, len(req.Context))
	for _, hit := range req.Context {
		citations = append(citations, rag.CitationFromHit(hit, hit.Chunk.Content))
	}
	return rag.GenerateResponse{
		Answer:       "Based on the indexed context: authentication is implemented in internal/auth/service.go.",
		InputTokens:  8,
		OutputTokens: 10,
		Model:        "test-llm",
		Citations:    citations,
	}, nil
}

type codeQATestRetriever struct {
	repositoryID uuid.UUID
	snapshotID   uuid.UUID
	hits         []rag.RetrievalHit
}

func (r codeQATestRetriever) Retrieve(_ context.Context, req rag.RetrievalRequest) (rag.RetrievalResult, error) {
	return rag.RetrievalResult{
		OriginalQuery:      req.Query,
		RewrittenQuery:     req.Query,
		RepositoryID:       r.repositoryID,
		SnapshotID:         r.snapshotID,
		BranchName:         req.BranchName,
		CommitSHA:          "abc123",
		DenseHits:          r.hits,
		FusedHits:          r.hits,
		RerankedHits:       r.hits,
		QueryCategory:      req.QueryCategory,
		RetrievalPath:      req.RetrievalPath,
		RetrievalConfig:    req.RetrievalConfig,
		RetrievedChunkIDs:  retrievedChunkIDs(r.hits),
		StageContributions: map[string]int{"dense": len(r.hits)},
		LatencyMS:          12,
	}, nil
}

type codeQAFollowUpRetriever struct {
	repositoryID uuid.UUID
	snapshotID   uuid.UUID
}

func (r codeQAFollowUpRetriever) Retrieve(_ context.Context, req rag.RetrievalRequest) (rag.RetrievalResult, error) {
	var hits []rag.RetrievalHit
	if strings.Contains(strings.ToLower(req.Query), "database") {
		hits = []rag.RetrievalHit{
			codeQATestHit(r.repositoryID, r.snapshotID, "internal/database/connect.go", "package database\nfunc Connect() {}\n// database connection setup uses pgx"),
		}
	} else {
		hits = []rag.RetrievalHit{
			codeQATestHit(r.repositoryID, r.snapshotID, "internal/auth/auth.go", "package auth\ntype JWTManager struct{}"),
		}
	}
	return rag.RetrievalResult{
		OriginalQuery:      req.Query,
		RewrittenQuery:     req.Query,
		RepositoryID:       r.repositoryID,
		SnapshotID:         r.snapshotID,
		BranchName:         req.BranchName,
		CommitSHA:          "abc123",
		DenseHits:          hits,
		FusedHits:          hits,
		RerankedHits:       hits,
		QueryCategory:      req.QueryCategory,
		RetrievalPath:      req.RetrievalPath,
		RetrievalConfig:    req.RetrievalConfig,
		RetrievedChunkIDs:  retrievedChunkIDs(hits),
		StageContributions: map[string]int{"dense": len(hits)},
		LatencyMS:          12,
	}, nil
}

type codeQAReportFollowUpRetriever struct {
	repositoryID uuid.UUID
	snapshotID   uuid.UUID
}

func (r codeQAReportFollowUpRetriever) Retrieve(_ context.Context, req rag.RetrievalRequest) (rag.RetrievalResult, error) {
	var hits []rag.RetrievalHit
	if strings.Contains(strings.ToLower(req.Query), "repository qa") {
		hit := codeQATestHit(r.repositoryID, r.snapshotID, "internal/codeqa/service.go", strings.Repeat("repository qa service report orchestration ", 1200))
		hit.Chunk.TokenCount = 3000
		hits = []rag.RetrievalHit{
			hit,
		}
	} else {
		hit := codeQATestHit(r.repositoryID, r.snapshotID, "internal/codeqa/reports.go", strings.Repeat("report section citation evidence quality ", 1200))
		hit.Chunk.TokenCount = 5000
		hits = []rag.RetrievalHit{hit}
	}
	return rag.RetrievalResult{
		OriginalQuery:      req.Query,
		RewrittenQuery:     req.Query,
		RepositoryID:       r.repositoryID,
		SnapshotID:         r.snapshotID,
		BranchName:         req.BranchName,
		CommitSHA:          "abc123",
		DenseHits:          hits,
		FusedHits:          hits,
		RerankedHits:       hits,
		QueryCategory:      req.QueryCategory,
		RetrievalPath:      req.RetrievalPath,
		RetrievalConfig:    req.RetrievalConfig,
		RetrievedChunkIDs:  retrievedChunkIDs(hits),
		StageContributions: map[string]int{"dense": len(hits)},
		LatencyMS:          12,
	}, nil
}

func codeQATestHit(repositoryID, snapshotID uuid.UUID, path string, content string) rag.RetrievalHit {
	return rag.RetrievalHit{
		Chunk: rag.Chunk{
			ID:       uuid.New(),
			VectorID: uuid.NewString(),
			Content:  content,
			Metadata: map[string]any{
				"repository_id": repositoryID.String(),
				"snapshot_id":   snapshotID.String(),
				"branch_name":   "main",
				"commit_sha":    "abc123",
				"path":          path,
				"start_line":    1,
				"end_line":      20,
			},
		},
		DenseScore: 0.9,
		Source:     "dense",
	}
}

func supportGateFromTrace(t *testing.T, trace repositories.RetrievalTraceInput) supportGateResult {
	t.Helper()
	config, ok := trace.RetrievalConfig.(map[string]any)
	if !ok {
		t.Fatalf("retrieval config = %#v", trace.RetrievalConfig)
	}
	gate, ok := config["support_gate"].(supportGateResult)
	if !ok {
		t.Fatalf("support gate = %#v", config["support_gate"])
	}
	return gate
}
