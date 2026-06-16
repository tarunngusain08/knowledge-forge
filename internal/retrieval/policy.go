package retrieval

import "strings"

const (
	CategoryExactLookup        = "exact_symbol_file_lookup"
	CategoryArchitecture       = "architecture_explanation"
	CategoryImplementation     = "implementation_question"
	CategoryImpact             = "impact_question"
	CategoryUnsupportedUnknown = "unsupported_unknown"
)

type PolicyOptions struct {
	TopK             int
	RerankerEnabled  bool
	HasSymbolIndex   bool
	HasGraphEvidence bool
}

type Policy struct {
	Category              string         `json:"category"`
	TopK                  int            `json:"top_k"`
	CandidateK            int            `json:"candidate_k"`
	ContextTokenBudget    int            `json:"context_token_budget"`
	RerankerEnabled       bool           `json:"reranker_enabled"`
	RerankerSkippedReason string         `json:"reranker_skipped_reason,omitempty"`
	RetrievalPath         []string       `json:"retrieval_path"`
	Config                map[string]any `json:"config"`
}

func BuildPolicy(question string, opts PolicyOptions) Policy {
	category := ClassifyQuestion(question)
	topK := opts.TopK
	if topK <= 0 {
		topK = 5
	}
	policy := Policy{
		Category:           category,
		TopK:               topK,
		CandidateK:         20,
		ContextTokenBudget: 2400,
		RerankerEnabled:    opts.RerankerEnabled,
		RetrievalPath:      []string{"dense"},
	}
	if opts.HasSymbolIndex {
		policy.RetrievalPath = append(policy.RetrievalPath, "symbol")
	}
	switch category {
	case CategoryExactLookup:
		policy.CandidateK = maxPositive(8, topK)
		policy.ContextTokenBudget = 1200
		policy.RerankerEnabled = false
		policy.RerankerSkippedReason = "exact lookup uses narrow candidate set"
	case CategoryArchitecture:
		policy.CandidateK = maxPositive(24, topK*4)
		policy.ContextTokenBudget = 3200
	case CategoryImpact:
		policy.CandidateK = maxPositive(30, topK*5)
		policy.ContextTokenBudget = 4000
		if opts.HasGraphEvidence {
			policy.RetrievalPath = append(policy.RetrievalPath, "graph")
		}
	case CategoryUnsupportedUnknown:
		policy.CandidateK = maxPositive(12, topK*2)
		policy.ContextTokenBudget = 1600
	default:
		policy.CandidateK = maxPositive(20, topK*4)
	}
	if policy.RerankerEnabled && policy.CandidateK <= policy.TopK {
		policy.RerankerEnabled = false
		policy.RerankerSkippedReason = "candidate count does not exceed final top_k"
	}
	if policy.RerankerEnabled {
		policy.RetrievalPath = append(policy.RetrievalPath, "rerank")
	}
	policy.Config = map[string]any{
		"category":                policy.Category,
		"top_k":                   policy.TopK,
		"candidate_k":             policy.CandidateK,
		"context_token_budget":    policy.ContextTokenBudget,
		"reranker_enabled":        policy.RerankerEnabled,
		"reranker_skipped_reason": policy.RerankerSkippedReason,
		"retrieval_path":          policy.RetrievalPath,
	}
	return policy
}

func ClassifyQuestion(question string) string {
	normalized := strings.ToLower(strings.TrimSpace(question))
	if normalized == "" {
		return CategoryUnsupportedUnknown
	}
	if containsAny(normalized, "impact", "affect", "affected", "break", "risk", "dependency", "depends on") {
		return CategoryImpact
	}
	if containsAny(normalized, "architecture", "design", "flow", "how does", "overview", "explain") {
		return CategoryArchitecture
	}
	if containsAny(normalized, "where is", "file", "symbol", "function", "struct", "interface", "class", ".go", ".ts", ".py") {
		return CategoryExactLookup
	}
	if containsAny(normalized, "implement", "implemented", "logic", "handles", "responsible", "add", "change") {
		return CategoryImplementation
	}
	if containsAny(normalized, "secret", "password", "token", "ignore instructions", "reveal") {
		return CategoryUnsupportedUnknown
	}
	return CategoryImplementation
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func maxPositive(a, b int) int {
	if a > b {
		return a
	}
	return b
}
