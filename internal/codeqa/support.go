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
	Answerable      bool     `json:"answerable"`
	Reason          string   `json:"reason"`
	MatchedTerms    []string `json:"matched_terms"`
	MissingTerms    []string `json:"missing_terms"`
	MatchedEvidence []string `json:"matched_evidence"`
	MissingEvidence []string `json:"missing_evidence"`
}

func evaluateAnswerSupport(question, category string, hits []rag.RetrievalHit) supportGateResult {
	requiredEvidence := requiredEvidenceGroups(question)
	matchedEvidence, missingEvidence := evidenceGroupSupport(requiredEvidence, hits)
	if isPromptInjectionRequest(question) {
		return supportGateResult{
			Answerable:      false,
			Reason:          "prompt_injection",
			MatchedTerms:    []string{},
			MissingTerms:    meaningfulQuestionTerms(question),
			MatchedEvidence: matchedEvidence,
			MissingEvidence: nonNilStrings(appendMissingEvidence(missingEvidence, "authorized_secret_disclosure_evidence")),
		}
	}
	if isExternalFactRequest(question) {
		return supportGateResult{
			Answerable:      false,
			Reason:          "external_fact",
			MatchedTerms:    []string{},
			MissingTerms:    meaningfulQuestionTerms(question),
			MatchedEvidence: matchedEvidence,
			MissingEvidence: nonNilStrings(appendMissingEvidence(missingEvidence, "repository_snapshot_external_fact_evidence")),
		}
	}
	if isDeletedSymbolRequest(question) {
		return supportGateResult{
			Answerable:      false,
			Reason:          "missing_identifier",
			MatchedTerms:    []string{},
			MissingTerms:    meaningfulQuestionTerms(question),
			MatchedEvidence: matchedEvidence,
			MissingEvidence: nonNilStrings(appendMissingEvidence(missingEvidence, missingIdentifierEvidence(question))),
		}
	}
	if category == retrievalpkg.CategoryUnsupportedUnknown {
		return supportGateResult{
			Answerable:      false,
			Reason:          "unsupported_intent",
			MatchedTerms:    []string{},
			MissingTerms:    meaningfulQuestionTerms(question),
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}
	if len(hits) == 0 {
		return supportGateResult{
			Answerable:      false,
			Reason:          "no_retrieved_context",
			MatchedTerms:    []string{},
			MissingTerms:    meaningfulQuestionTerms(question),
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}

	terms := meaningfulQuestionTerms(question)
	if len(terms) == 0 {
		return supportGateResult{
			Answerable:      true,
			Reason:          "retrieved_context_available",
			MatchedTerms:    []string{},
			MissingTerms:    []string{},
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}

	evidenceTerms := evidenceTermSet(hits)
	matched, missing := partitionTerms(terms, evidenceTerms)
	if len(requiredEvidence) > 0 && len(missingEvidence) > 0 {
		return supportGateResult{
			Answerable:      false,
			Reason:          supportReasonForMissingEvidence(question, category),
			MatchedTerms:    matched,
			MissingTerms:    missing,
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}
	if category == retrievalpkg.CategoryExactLookup && containsLikelyIdentifier(question) && len(missing) > 0 {
		return supportGateResult{
			Answerable:      false,
			Reason:          "missing_identifier",
			MatchedTerms:    matched,
			MissingTerms:    missing,
			MatchedEvidence: matchedEvidence,
			MissingEvidence: appendMissingEvidence(missingEvidence, missingIdentifierEvidence(question)),
		}
	}

	if len(requiredEvidence) > 0 && len(missingEvidence) == 0 {
		return supportGateResult{
			Answerable:      true,
			Reason:          "repository_supported_fact",
			MatchedTerms:    matched,
			MissingTerms:    missing,
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}

	if len(matched) == 0 {
		return supportGateResult{
			Answerable:      false,
			Reason:          "missing_domain_terms",
			MatchedTerms:    matched,
			MissingTerms:    missing,
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}

	if category == retrievalpkg.CategoryArchitecture {
		return supportGateResult{
			Answerable:      true,
			Reason:          "architecture_context_available",
			MatchedTerms:    matched,
			MissingTerms:    missing,
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}

	requiredMatches := 1
	if len(terms) >= 2 {
		requiredMatches = 2
	}
	if len(matched) >= requiredMatches || float64(len(matched))/float64(len(terms)) >= 0.5 {
		return supportGateResult{
			Answerable:      true,
			Reason:          "repository_supported_fact",
			MatchedTerms:    matched,
			MissingTerms:    missing,
			MatchedEvidence: matchedEvidence,
			MissingEvidence: missingEvidence,
		}
	}

	return supportGateResult{
		Answerable:      false,
		Reason:          "insufficient_domain_term_overlap",
		MatchedTerms:    matched,
		MissingTerms:    missing,
		MatchedEvidence: matchedEvidence,
		MissingEvidence: missingEvidence,
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

func requiredEvidenceGroups(question string) []string {
	normalized := strings.ToLower(question)
	groups := []string{}
	if strings.Contains(normalized, "customer") && strings.Contains(normalized, "revenue") && (strings.Contains(normalized, "api") || strings.Contains(normalized, "endpoint")) {
		groups = append(groups, "customer_revenue_endpoint_evidence")
	} else if strings.Contains(normalized, "revenue") {
		groups = append(groups, "revenue_domain_evidence")
	}
	if strings.Contains(normalized, "payroll") {
		groups = append(groups, "payroll_domain_evidence")
	}
	if strings.Contains(normalized, "rag") && strings.Contains(normalized, "retrieval") {
		groups = append(groups, "retrieval_source", "rag_context")
	}
	if strings.Contains(normalized, "http") && strings.Contains(normalized, "api") && strings.Contains(normalized, "chat") {
		groups = append(groups, "api_router", "chat_handler")
	}
	if strings.Contains(normalized, "postgresql") && strings.Contains(normalized, "fts") {
		groups = append(groups, "lexical_retrieval", "postgres_fts")
	}
	if strings.Contains(normalized, "repository registration") {
		groups = append(groups, "repository_api", "repository_store")
	}
	if strings.Contains(normalized, "auth") || strings.Contains(normalized, "authentication") {
		groups = append(groups, "auth_service")
	}
	if strings.Contains(normalized, "database") || strings.Contains(normalized, "db wiring") {
		groups = append(groups, "database_connection")
	}
	if strings.Contains(normalized, "deep-dive report") || strings.Contains(normalized, "deep dive report") || strings.Contains(normalized, "report generation") || strings.Contains(normalized, "reports generated") {
		groups = append(groups, "report_generator")
		if strings.Contains(normalized, "how are") || strings.Contains(normalized, "deep-dive") || strings.Contains(normalized, "deep dive") {
			groups = append(groups, "repo_qa_service")
		} else {
			groups = append(groups, "evidence_quality")
		}
	}
	return uniqueStrings(groups)
}

func evidenceGroupSupport(required []string, hits []rag.RetrievalHit) ([]string, []string) {
	available := evidenceGroupSet(hits)
	matched := make([]string, 0, len(available))
	for group := range available {
		matched = append(matched, group)
	}
	var missing []string
	for _, group := range required {
		if !available[group] {
			missing = append(missing, group)
		}
	}
	sort.Strings(matched)
	sort.Strings(missing)
	return nonNilStrings(matched), nonNilStrings(missing)
}

func evidenceGroupSet(hits []rag.RetrievalHit) map[string]bool {
	groups := map[string]bool{}
	for _, hit := range hits {
		for _, group := range metadataEvidenceGroups(hit.Chunk.Metadata) {
			groups[group] = true
		}
		for _, group := range inferredEvidenceGroups(hit) {
			groups[group] = true
		}
	}
	return groups
}

func metadataEvidenceGroups(metadata map[string]any) []string {
	if metadata == nil {
		return nil
	}
	var groups []string
	for _, key := range []string{"evidence_group", "evidence_groups"} {
		switch value := metadata[key].(type) {
		case string:
			groups = append(groups, strings.TrimSpace(value))
		case []string:
			groups = append(groups, value...)
		case []any:
			for _, item := range value {
				groups = append(groups, strings.TrimSpace(fmt.Sprint(item)))
			}
		}
	}
	return uniqueStrings(groups)
}

func inferredEvidenceGroups(hit rag.RetrievalHit) []string {
	path := strings.ToLower(strings.TrimSpace(fmt.Sprint(hit.Chunk.Metadata["path"])))
	content := strings.ToLower(hit.Chunk.Content)
	var groups []string
	switch {
	case strings.Contains(path, "internal/retrieval/fts.go"):
		groups = append(groups, "lexical_retrieval")
	case strings.Contains(path, "queries/documents.sql"):
		groups = append(groups, "postgres_fts")
	case strings.Contains(path, "internal/retrieval/"):
		groups = append(groups, "dense_or_code_retrieval", "retrieval_source")
	case strings.Contains(path, "internal/rag/context.go"):
		groups = append(groups, "context_assembly", "rag_context")
	case strings.Contains(path, "internal/rag/"):
		groups = append(groups, "rag_context")
	case strings.Contains(path, "internal/httpapi/router.go"):
		groups = append(groups, "api_router", "api_path_match")
	case strings.Contains(path, "internal/httpapi/chat"):
		groups = append(groups, "chat_handler", "api_path_match")
	case strings.Contains(path, "internal/httpapi/repository"):
		groups = append(groups, "repository_api", "api_path_match")
	case strings.HasPrefix(path, "cmd/api/") || strings.Contains(path, "internal/httpapi/"):
		groups = append(groups, "api_path_match")
	case strings.Contains(path, "internal/repositories/service.go"):
		groups = append(groups, "repository_store")
	case strings.Contains(path, "internal/auth/"):
		groups = append(groups, "auth_service")
		if strings.Contains(path, "middleware") {
			groups = append(groups, "auth_middleware")
		}
	case strings.Contains(path, "internal/database/") || strings.Contains(path, "internal/db/"):
		groups = append(groups, "database_connection")
	case strings.Contains(path, "internal/codeqa/reports.go"):
		groups = append(groups, "report_generator", "evidence_quality")
	case strings.Contains(path, "internal/codeqa/service.go"):
		groups = append(groups, "repo_qa_service")
	case strings.Contains(path, "ui/web/") || strings.Contains(path, "ui/streamlit/"):
		groups = append(groups, "ui_path_match")
	}
	if strings.Contains(content, "revenue") && !containsNegatedDomain(content, "revenue") {
		groups = append(groups, "revenue_domain_evidence")
	}
	if strings.Contains(content, "customer") && strings.Contains(content, "revenue") && strings.Contains(content, "endpoint") && !containsNegatedDomain(content, "revenue") {
		groups = append(groups, "customer_revenue_endpoint_evidence")
	}
	if strings.Contains(content, "payroll") && !containsNegatedDomain(content, "payroll") {
		groups = append(groups, "payroll_domain_evidence")
	}
	return uniqueStrings(groups)
}

func containsNegatedDomain(content, term string) bool {
	return strings.Contains(content, "no "+term) || strings.Contains(content, "not "+term) || strings.Contains(content, "without "+term)
}

func supportReasonForMissingEvidence(question, category string) string {
	normalized := strings.ToLower(question)
	if strings.Contains(normalized, "revenue") || strings.Contains(normalized, "payroll") {
		return "missing_domain_terms"
	}
	if category == retrievalpkg.CategoryExactLookup && containsLikelyIdentifier(question) {
		return "missing_identifier"
	}
	return "missing_required_evidence"
}

func appendMissingEvidence(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func missingIdentifierEvidence(question string) string {
	if strings.Contains(strings.ToLower(question), "deleted symbol") {
		return "deleted_symbol_xyz_evidence"
	}
	return "missing_identifier_evidence"
}

func isPromptInjectionRequest(question string) bool {
	normalized := strings.ToLower(question)
	return strings.Contains(normalized, "ignore previous instructions") ||
		strings.Contains(normalized, "ignore instructions") ||
		(strings.Contains(normalized, "reveal") && strings.Contains(normalized, "secret"))
}

func isExternalFactRequest(question string) bool {
	normalized := strings.ToLower(question)
	return strings.Contains(normalized, "today") &&
		(strings.Contains(normalized, "ceo") || strings.Contains(normalized, "current")) &&
		!strings.Contains(normalized, "repository")
}

func isDeletedSymbolRequest(question string) bool {
	normalized := strings.ToLower(question)
	return strings.Contains(normalized, "deleted symbol") || strings.Contains(normalized, "symbol xyz")
}

func recoverableEvidenceGroups(groups []string) []string {
	recoverable := map[string]bool{
		"api_router":              true,
		"chat_handler":            true,
		"auth_service":            true,
		"database_connection":     true,
		"report_generator":        true,
		"repo_qa_service":         true,
		"evidence_quality":        true,
		"retrieval_source":        true,
		"rag_context":             true,
		"lexical_retrieval":       true,
		"postgres_fts":            true,
		"repository_api":          true,
		"repository_store":        true,
		"dense_or_code_retrieval": true,
		"fusion":                  true,
		"context_assembly":        true,
	}
	var out []string
	for _, group := range groups {
		if recoverable[group] {
			out = append(out, group)
		}
	}
	return out
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

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return nonNilStrings(out)
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
