package codeqa

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
	retrievalpkg "github.com/tarunngusain08/knowledge-forge/internal/retrieval"
)

const refusalAnswer = "I could not find this in the indexed context."

type supportGateResult struct {
	Answerable   bool     `json:"answerable"`
	Reason       string   `json:"reason"`
	MatchedTerms []string `json:"matched_terms"`
	MissingTerms []string `json:"missing_terms"`
}

func evaluateAnswerSupport(question, category string, hits []rag.RetrievalHit) supportGateResult {
	if category == retrievalpkg.CategoryUnsupportedUnknown {
		return supportGateResult{
			Answerable:   false,
			Reason:       "unsupported_intent",
			MatchedTerms: []string{},
			MissingTerms: meaningfulQuestionTerms(question),
		}
	}
	if len(hits) == 0 {
		return supportGateResult{
			Answerable:   false,
			Reason:       "no_retrieved_context",
			MatchedTerms: []string{},
			MissingTerms: meaningfulQuestionTerms(question),
		}
	}

	terms := meaningfulQuestionTerms(question)
	if len(terms) == 0 {
		return supportGateResult{
			Answerable:   true,
			Reason:       "retrieved_context_available",
			MatchedTerms: []string{},
			MissingTerms: []string{},
		}
	}

	evidenceTerms := evidenceTermSet(hits)
	matched, missing := partitionTerms(terms, evidenceTerms)
	if len(matched) == 0 {
		return supportGateResult{
			Answerable:   false,
			Reason:       "missing_domain_terms",
			MatchedTerms: matched,
			MissingTerms: missing,
		}
	}

	if category == retrievalpkg.CategoryArchitecture {
		return supportGateResult{
			Answerable:   true,
			Reason:       "architecture_context_available",
			MatchedTerms: matched,
			MissingTerms: missing,
		}
	}

	if category == retrievalpkg.CategoryExactLookup && containsLikelyIdentifier(question) && len(missing) > 0 {
		return supportGateResult{
			Answerable:   false,
			Reason:       "missing_exact_identifier_terms",
			MatchedTerms: matched,
			MissingTerms: missing,
		}
	}

	requiredMatches := 1
	if len(terms) >= 2 {
		requiredMatches = 2
	}
	if len(matched) >= requiredMatches || float64(len(matched))/float64(len(terms)) >= 0.5 {
		return supportGateResult{
			Answerable:   true,
			Reason:       "domain_terms_supported",
			MatchedTerms: matched,
			MissingTerms: missing,
		}
	}

	return supportGateResult{
		Answerable:   false,
		Reason:       "insufficient_domain_term_overlap",
		MatchedTerms: matched,
		MissingTerms: missing,
	}
}

func retrievalConfigWithSupportGate(config map[string]any, gate supportGateResult) map[string]any {
	out := map[string]any{}
	for key, value := range config {
		out[key] = value
	}
	out["support_gate"] = gate
	return out
}

func meaningfulQuestionTerms(question string) []string {
	seen := map[string]bool{}
	var terms []string
	for _, token := range tokenizeForSupport(question) {
		token = normalizeSupportToken(token)
		if token == "" || supportStopWords[token] || seen[token] {
			continue
		}
		seen[token] = true
		terms = append(terms, token)
	}
	return terms
}

func evidenceTermSet(hits []rag.RetrievalHit) map[string]bool {
	terms := map[string]bool{}
	for _, hit := range hits {
		addSupportTerms(terms, hit.Chunk.Content)
		for _, value := range hit.Chunk.Metadata {
			addSupportTerms(terms, fmt.Sprint(value))
		}
	}
	return terms
}

func addSupportTerms(out map[string]bool, text string) {
	for _, token := range tokenizeForSupport(text) {
		token = normalizeSupportToken(token)
		if token == "" {
			continue
		}
		out[token] = true
	}
}

func partitionTerms(terms []string, evidence map[string]bool) ([]string, []string) {
	var matched []string
	var missing []string
	for _, term := range terms {
		if evidence[term] {
			matched = append(matched, term)
			continue
		}
		missing = append(missing, term)
	}
	sort.Strings(matched)
	sort.Strings(missing)
	return nonNilStrings(matched), nonNilStrings(missing)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func tokenizeForSupport(text string) []string {
	var builder strings.Builder
	var previous rune
	for _, r := range text {
		if previous != 0 && unicode.IsLower(previous) && unicode.IsUpper(r) {
			builder.WriteRune(' ')
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(unicode.ToLower(r))
		default:
			builder.WriteRune(' ')
		}
		previous = r
	}
	return strings.Fields(builder.String())
}

func normalizeSupportToken(token string) string {
	switch token {
	case "authentication", "authenticate", "authenticating", "authenticated":
		return "auth"
	case "authorization", "authorize", "authorized":
		return "authz"
	case "charged", "charges", "charging":
		return "charge"
	case "implemented", "implementing", "implementation":
		return "implement"
	case "processes", "processed", "processing":
		return "process"
	}
	if len(token) > 4 && strings.HasSuffix(token, "ies") {
		return token[:len(token)-3] + "y"
	}
	if len(token) > 4 && strings.HasSuffix(token, "es") {
		return token[:len(token)-2]
	}
	if len(token) > 3 && strings.HasSuffix(token, "s") {
		return token[:len(token)-1]
	}
	return token
}

func containsLikelyIdentifier(question string) bool {
	for _, token := range strings.Fields(question) {
		hasLower := false
		hasInnerUpper := false
		for idx, r := range token {
			hasLower = hasLower || unicode.IsLower(r)
			hasInnerUpper = hasInnerUpper || (idx > 0 && unicode.IsUpper(r))
		}
		if hasLower && hasInnerUpper {
			return true
		}
	}
	return strings.Contains(question, ".go") || strings.Contains(question, ".ts") || strings.Contains(question, ".py")
}

var supportStopWords = map[string]bool{
	"a": true, "about": true, "across": true, "all": true, "an": true, "and": true, "are": true, "as": true,
	"at": true, "available": true, "be": true, "been": true, "being": true, "broad": true, "by": true, "can": true,
	"claim": true, "cited": true, "context": true, "could": true, "describe": true, "detail": true, "details": true,
	"did": true, "do": true, "does": true, "due": true, "enough": true, "entry": true, "every": true, "evidence": true,
	"explain": true, "file": true, "files": true, "find": true, "first": true, "for": true, "from": true, "generate": true,
	"generated": true, "handle": true, "handles": true, "how": true, "i": true, "identify": true, "if": true, "in": true,
	"indexed": true, "instead": true, "insufficient": true, "into": true, "invent": true, "is": true, "layer": true,
	"locate": true, "main": true, "major": true, "me": true, "missing": true, "must": true, "no": true, "of": true,
	"on": true, "only": true, "or": true, "our": true, "pieces": true, "point": true, "points": true, "question": true,
	"report": true, "repository": true, "request": true, "retrieved": true, "say": true, "section": true, "service": true,
	"show": true, "state": true, "supported": true, "system": true, "task": true, "tell": true, "that": true, "the": true,
	"this": true, "to": true, "use": true, "was": true, "were": true, "what": true, "when": true, "where": true,
	"which": true, "who": true, "why": true, "with": true, "would": true,
}
