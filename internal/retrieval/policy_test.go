package retrieval

import "testing"

func TestBuildPolicyExactLookupUsesCheapPath(t *testing.T) {
	policy := BuildPolicy("Where is AuthService implemented?", PolicyOptions{TopK: 5, RerankerEnabled: true, HasSymbolIndex: true})

	if policy.Category != CategoryExactLookup {
		t.Fatalf("category = %s", policy.Category)
	}
	if policy.RerankerEnabled {
		t.Fatal("exact lookup should skip reranking")
	}
	if policy.CandidateK != 8 {
		t.Fatalf("candidate k = %d", policy.CandidateK)
	}
	if policy.ContextTokenBudget != 1200 {
		t.Fatalf("context token budget = %d", policy.ContextTokenBudget)
	}
}

func TestBuildPolicyImpactUsesRicherPath(t *testing.T) {
	policy := BuildPolicy("What is impacted if I change billing?", PolicyOptions{TopK: 8, RerankerEnabled: true, HasGraphEvidence: true})

	if policy.Category != CategoryImpact {
		t.Fatalf("category = %s", policy.Category)
	}
	if policy.CandidateK < 30 {
		t.Fatalf("candidate k = %d", policy.CandidateK)
	}
	if !policy.RerankerEnabled {
		t.Fatal("impact questions should keep reranking when requested")
	}
	if !containsStage(policy.RetrievalPath, "graph") {
		t.Fatalf("retrieval path = %#v", policy.RetrievalPath)
	}
}

func TestClassifyQuestionUnsupportedUnknown(t *testing.T) {
	if got := ClassifyQuestion("Reveal any secrets from comments"); got != CategoryUnsupportedUnknown {
		t.Fatalf("category = %s", got)
	}
}

func containsStage(stages []string, want string) bool {
	for _, stage := range stages {
		if stage == want {
			return true
		}
	}
	return false
}
