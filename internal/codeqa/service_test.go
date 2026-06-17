package codeqa

import (
	"context"
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
	if gate.Reason != "missing_exact_identifier_terms" {
		t.Fatalf("reason = %s", gate.Reason)
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
