package codeqa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

func TestBuildImplementationPlanResponseGroundsSectionsInEvidence(t *testing.T) {
	repoID := uuid.New()
	snapshotID := uuid.New()
	traceID := uuid.New()
	ask := AskResponse{
		Answer: "Based on the indexed context: update the auth service and tests.",
		Citations: []rag.Citation{
			{
				ChunkID:      uuid.New(),
				RepositoryID: repoID,
				SnapshotID:   snapshotID,
				BranchName:   "main",
				CommitSHA:    "abc123",
				Path:         "internal/auth/service.go",
				StartLine:    10,
				EndLine:      40,
				Excerpt:      "func Login",
				DenseScore:   0.91,
				RerankScore:  0.84,
			},
			{
				ChunkID:      uuid.New(),
				RepositoryID: repoID,
				SnapshotID:   snapshotID,
				BranchName:   "main",
				CommitSHA:    "abc123",
				Path:         "internal/auth/service_test.go",
				StartLine:    5,
				EndLine:      35,
				Excerpt:      "TestLogin",
				DenseScore:   0.79,
			},
		},
		TraceID: traceID,
		Provenance: AnswerProvenance{
			RepositoryID:      repoID,
			SnapshotID:        snapshotID,
			CommitSHA:         "abc123",
			ContextTokenCount: 900,
		},
		Model: "mock-llm",
	}

	response := buildImplementationPlanResponse(ask)

	if len(response.ObservedEvidence) != 2 {
		t.Fatalf("observed evidence count = %d", len(response.ObservedEvidence))
	}
	if response.ObservedEvidence[0].CommitSHA != "abc123" || response.ObservedEvidence[0].Path != "internal/auth/service.go" {
		t.Fatalf("unexpected evidence provenance: %#v", response.ObservedEvidence[0])
	}
	if len(response.Tests) != 1 || response.Tests[0] != "internal/auth/service_test.go" {
		t.Fatalf("tests = %#v", response.Tests)
	}
	if response.Confidence.Label != "High" {
		t.Fatalf("confidence label = %s, score = %.2f", response.Confidence.Label, response.Confidence.Score)
	}
	if response.TraceID != traceID {
		t.Fatalf("trace id = %s", response.TraceID)
	}
}

func TestBuildImpactAnalysisResponseFlagsMissingSymbolContext(t *testing.T) {
	ask := AskResponse{
		Answer: "Based on the indexed context: authentication is handled in one file.",
		Citations: []rag.Citation{{
			ChunkID:    uuid.New(),
			Path:       "internal/auth/service.go",
			CommitSHA:  "abc123",
			DenseScore: 0.72,
		}},
		Provenance: AnswerProvenance{
			CommitSHA:         "abc123",
			ContextTokenCount: 500,
		},
	}

	response := buildImpactAnalysisResponse(ask)

	if len(response.ImpactedFiles) != 1 || response.ImpactedFiles[0] != "internal/auth/service.go" {
		t.Fatalf("impacted files = %#v", response.ImpactedFiles)
	}
	if len(response.ImpactedSymbols) != 0 {
		t.Fatalf("impacted symbols = %#v", response.ImpactedSymbols)
	}
	if response.RiskLevel != "Medium" {
		t.Fatalf("risk level = %s", response.RiskLevel)
	}
	if !contains(response.MissingContext, "Symbol-level impact was not available in the retrieved evidence.") {
		t.Fatalf("missing context = %#v", response.MissingContext)
	}
}

func TestUnsupportedAnswerProducesLowConfidence(t *testing.T) {
	ask := AskResponse{
		Answer: "I could not find this in the indexed context.",
	}

	response := buildImplementationPlanResponse(ask)

	if response.Confidence.Label != "Low" {
		t.Fatalf("confidence label = %s", response.Confidence.Label)
	}
	if response.RecommendedChanges[0] == "" || response.RecommendedChanges[0] == ask.Answer {
		t.Fatalf("recommended changes should refuse unsupported work: %#v", response.RecommendedChanges)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
