package codeqa

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

type WorkflowRequest struct {
	UserID          uuid.UUID `json:"user_id"`
	RepositoryID    uuid.UUID `json:"repository_id"`
	BranchName      string    `json:"branch_name"`
	Request         string    `json:"request"`
	TopK            int       `json:"top_k"`
	RerankerEnabled *bool     `json:"reranker_enabled,omitempty"`
}

type EvidenceItem struct {
	ChunkID      uuid.UUID `json:"chunk_id"`
	RepositoryID uuid.UUID `json:"repository_id"`
	SnapshotID   uuid.UUID `json:"snapshot_id"`
	BranchName   string    `json:"branch_name"`
	CommitSHA    string    `json:"commit_sha"`
	Path         string    `json:"path"`
	StartLine    int       `json:"start_line"`
	EndLine      int       `json:"end_line"`
	Excerpt      string    `json:"excerpt"`
	DenseScore   float64   `json:"dense_score,omitempty"`
	RerankScore  float64   `json:"rerank_score,omitempty"`
	Reasons      []string  `json:"reasons,omitempty"`
}

type EvidenceConfidence struct {
	Label             string   `json:"label"`
	Score             float64  `json:"score"`
	EvidenceCoverage  float64  `json:"evidence_coverage"`
	CitationCount     int      `json:"citation_count"`
	ContextTokenCount int      `json:"context_token_count"`
	Reasons           []string `json:"reasons"`
}

type ImplementationPlanResponse struct {
	ObservedEvidence   []EvidenceItem     `json:"observed_evidence"`
	RecommendedChanges []string           `json:"recommended_changes"`
	Assumptions        []string           `json:"assumptions"`
	MissingContext     []string           `json:"missing_context"`
	Risks              []string           `json:"risks"`
	Tests              []string           `json:"tests"`
	Confidence         EvidenceConfidence `json:"confidence"`
	Answer             string             `json:"answer"`
	Citations          []rag.Citation     `json:"citations"`
	TraceID            uuid.UUID          `json:"trace_id"`
	Provenance         AnswerProvenance   `json:"provenance"`
	Model              string             `json:"model"`
}

type ImpactAnalysisResponse struct {
	ObservedEvidence    []EvidenceItem     `json:"observed_evidence"`
	ImpactedFiles       []string           `json:"impacted_files"`
	ImpactedSymbols     []string           `json:"impacted_symbols"`
	AffectedTests       []string           `json:"affected_tests"`
	DependencyReasoning []string           `json:"dependency_reasoning"`
	RiskLevel           string             `json:"risk_level"`
	MissingContext      []string           `json:"missing_context"`
	Confidence          EvidenceConfidence `json:"confidence"`
	Answer              string             `json:"answer"`
	Citations           []rag.Citation     `json:"citations"`
	TraceID             uuid.UUID          `json:"trace_id"`
	Provenance          AnswerProvenance   `json:"provenance"`
	Model               string             `json:"model"`
}

func (s *Service) GenerateImplementationPlan(ctx context.Context, req WorkflowRequest) (ImplementationPlanResponse, error) {
	ask, err := s.Ask(ctx, AskRequest{
		UserID:          req.UserID,
		RepositoryID:    req.RepositoryID,
		BranchName:      req.BranchName,
		Question:        implementationPlanPrompt(req.Request),
		TopK:            workflowTopK(req.TopK),
		RerankerEnabled: req.RerankerEnabled,
	})
	if err != nil {
		return ImplementationPlanResponse{}, err
	}
	return buildImplementationPlanResponse(ask), nil
}

func (s *Service) AnalyzeImpact(ctx context.Context, req WorkflowRequest) (ImpactAnalysisResponse, error) {
	ask, err := s.Ask(ctx, AskRequest{
		UserID:          req.UserID,
		RepositoryID:    req.RepositoryID,
		BranchName:      req.BranchName,
		Question:        impactAnalysisPrompt(req.Request),
		TopK:            workflowTopK(req.TopK),
		RerankerEnabled: req.RerankerEnabled,
	})
	if err != nil {
		return ImpactAnalysisResponse{}, err
	}
	return buildImpactAnalysisResponse(ask), nil
}

func buildImplementationPlanResponse(ask AskResponse) ImplementationPlanResponse {
	evidence := evidenceItems(ask.Citations)
	missing := missingContext(ask, false)
	return ImplementationPlanResponse{
		ObservedEvidence:   evidence,
		RecommendedChanges: recommendedChanges(ask),
		Assumptions:        assumptions(ask),
		MissingContext:     missing,
		Risks:              planningRisks(ask),
		Tests:              testFiles(ask.Citations),
		Confidence:         confidenceFromEvidence(ask, missing),
		Answer:             ask.Answer,
		Citations:          ask.Citations,
		TraceID:            ask.TraceID,
		Provenance:         ask.Provenance,
		Model:              ask.Model,
	}
}

func buildImpactAnalysisResponse(ask AskResponse) ImpactAnalysisResponse {
	missing := missingContext(ask, true)
	files := citationFiles(ask.Citations)
	tests := testFiles(ask.Citations)
	return ImpactAnalysisResponse{
		ObservedEvidence:    evidenceItems(ask.Citations),
		ImpactedFiles:       files,
		ImpactedSymbols:     citationSymbols(ask.Citations),
		AffectedTests:       tests,
		DependencyReasoning: dependencyReasoning(files, tests),
		RiskLevel:           riskLevel(ask, tests, missing),
		MissingContext:      missing,
		Confidence:          confidenceFromEvidence(ask, missing),
		Answer:              ask.Answer,
		Citations:           ask.Citations,
		TraceID:             ask.TraceID,
		Provenance:          ask.Provenance,
		Model:               ask.Model,
	}
}

func implementationPlanPrompt(request string) string {
	if strings.TrimSpace(request) == "" {
		request = "Create an implementation plan for the requested repository change."
	}
	return fmt.Sprintf(`Implementation planning request:
%s

Return an evidence-grounded implementation plan. Use only the retrieved repository context.
Required sections: Observed Evidence, Recommended Changes, Assumptions, Missing Context, Risks, Tests, Confidence.
Every recommendation must be traceable to cited files or must be placed under Assumptions or Missing Context.`, strings.TrimSpace(request))
}

func impactAnalysisPrompt(request string) string {
	if strings.TrimSpace(request) == "" {
		request = "Analyze the likely impact of the requested repository change."
	}
	return fmt.Sprintf(`Impact analysis request:
%s

Return an evidence-grounded impact analysis. Use only the retrieved repository context.
Required sections: Observed Evidence, Impacted Files, Impacted Symbols, Affected Tests, Dependency Reasoning, Risk Level, Missing Context, Confidence.
If the retrieved evidence is insufficient, state uncertainty instead of inventing dependencies.`, strings.TrimSpace(request))
}

func workflowTopK(topK int) int {
	if topK <= 0 {
		return 8
	}
	return topK
}

func evidenceItems(citations []rag.Citation) []EvidenceItem {
	items := make([]EvidenceItem, 0, len(citations))
	for _, citation := range citations {
		items = append(items, EvidenceItem{
			ChunkID:      citation.ChunkID,
			RepositoryID: citation.RepositoryID,
			SnapshotID:   citation.SnapshotID,
			BranchName:   citation.BranchName,
			CommitSHA:    citation.CommitSHA,
			Path:         citation.Path,
			StartLine:    citation.StartLine,
			EndLine:      citation.EndLine,
			Excerpt:      citation.Excerpt,
			DenseScore:   citation.DenseScore,
			RerankScore:  citation.RerankScore,
			Reasons:      metadataReasons(citation.Metadata),
		})
	}
	return items
}

func recommendedChanges(ask AskResponse) []string {
	if len(ask.Citations) == 0 || unsupportedAnswer(ask.Answer) {
		return []string{"No implementation change is recommended until repository evidence is available."}
	}
	return []string{strings.TrimSpace(ask.Answer)}
}

func assumptions(ask AskResponse) []string {
	if len(ask.Citations) == 0 {
		return []string{"No repository evidence was retrieved, so implementation details would be speculative."}
	}
	return []string{"Recommendations are limited to the cited repository snapshot and do not assume behavior outside retrieved evidence."}
}

func missingContext(ask AskResponse, needsSymbols bool) []string {
	var missing []string
	if len(ask.Citations) == 0 || unsupportedAnswer(ask.Answer) {
		missing = append(missing, "No cited repository context was retrieved for this request.")
	}
	if ask.Provenance.CommitSHA == "" {
		missing = append(missing, "Repository commit SHA is unavailable, so the answer is not fully reproducible.")
	}
	if needsSymbols && len(citationSymbols(ask.Citations)) == 0 {
		missing = append(missing, "Symbol-level impact was not available in the retrieved evidence.")
	}
	if len(testFiles(ask.Citations)) == 0 {
		missing = append(missing, "No affected tests were identified from the retrieved citations.")
	}
	if len(missing) == 0 {
		missing = append(missing, "No missing context was detected from retrieval metadata.")
	}
	return missing
}

func planningRisks(ask AskResponse) []string {
	if len(ask.Citations) == 0 {
		return []string{"High hallucination risk because no cited evidence was retrieved."}
	}
	risks := []string{"Plan may miss code paths that were not retrieved or indexed."}
	if len(testFiles(ask.Citations)) == 0 {
		risks = append(risks, "Test impact is uncertain because no cited test files were retrieved.")
	}
	return risks
}

func dependencyReasoning(files, tests []string) []string {
	if len(files) == 0 {
		return []string{"No dependency reasoning is available because no impacted files were cited."}
	}
	reasoning := []string{"Impact is inferred from cited file evidence in the current repository snapshot."}
	if len(tests) == 0 {
		reasoning = append(reasoning, "No affected tests were retrieved, so test coverage impact is uncertain.")
	}
	return reasoning
}

func riskLevel(ask AskResponse, tests, missing []string) string {
	switch {
	case len(ask.Citations) == 0 || unsupportedAnswer(ask.Answer):
		return "High"
	case len(tests) == 0 || len(missing) > 1:
		return "Medium"
	default:
		return "Low"
	}
}

func confidenceFromEvidence(ask AskResponse, missing []string) EvidenceConfidence {
	citationCount := len(ask.Citations)
	coverage := math.Min(1, float64(citationCount)/3)
	score := 0.10
	reasons := []string{}
	if citationCount > 0 {
		score += 0.35
		reasons = append(reasons, "cited evidence was retrieved")
	}
	if len(citationFiles(ask.Citations)) > 1 {
		score += 0.15
		reasons = append(reasons, "multiple files support the response")
	}
	if ask.Provenance.ContextTokenCount >= 400 {
		score += 0.15
		reasons = append(reasons, "context budget contains enough evidence for synthesis")
	}
	if maxRetrievalScore(ask.Citations) >= 0.5 {
		score += 0.15
		reasons = append(reasons, "retrieval scores are strong")
	}
	if ask.Provenance.CommitSHA != "" {
		score += 0.10
		reasons = append(reasons, "answer is tied to a reproducible commit SHA")
	}
	if len(missing) > 1 {
		score -= 0.15
		reasons = append(reasons, "missing context reduces confidence")
	}
	score = clamp(score, 0, 1)
	label := "Low"
	if score >= 0.75 {
		label = "High"
	} else if score >= 0.45 {
		label = "Medium"
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "confidence is low because evidence is unavailable")
	}
	return EvidenceConfidence{
		Label:             label,
		Score:             round2(score),
		EvidenceCoverage:  round2(coverage),
		CitationCount:     citationCount,
		ContextTokenCount: ask.Provenance.ContextTokenCount,
		Reasons:           reasons,
	}
}

func citationFiles(citations []rag.Citation) []string {
	seen := map[string]bool{}
	var files []string
	for _, citation := range citations {
		path := strings.TrimSpace(citation.Path)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

func citationSymbols(citations []rag.Citation) []string {
	seen := map[string]bool{}
	var symbols []string
	for _, citation := range citations {
		for _, key := range []string{"symbol", "symbol_name", "qualified_name"} {
			value := strings.TrimSpace(fmt.Sprint(citation.Metadata[key]))
			if value == "" || value == "<nil>" || seen[value] {
				continue
			}
			seen[value] = true
			symbols = append(symbols, value)
		}
	}
	sort.Strings(symbols)
	return symbols
}

func testFiles(citations []rag.Citation) []string {
	seen := map[string]bool{}
	var tests []string
	for _, file := range citationFiles(citations) {
		lower := strings.ToLower(file)
		if !(strings.Contains(lower, "_test.") || strings.Contains(lower, "/test/") || strings.HasSuffix(lower, ".test.ts") || strings.HasSuffix(lower, ".spec.ts")) {
			continue
		}
		if seen[file] {
			continue
		}
		seen[file] = true
		tests = append(tests, file)
	}
	return tests
}

func maxRetrievalScore(citations []rag.Citation) float64 {
	maxScore := 0.0
	for _, citation := range citations {
		if citation.RerankScore > maxScore {
			maxScore = citation.RerankScore
		}
		if citation.DenseScore > maxScore {
			maxScore = citation.DenseScore
		}
	}
	return maxScore
}

func metadataReasons(metadata map[string]any) []string {
	if metadata == nil {
		return nil
	}
	value, ok := metadata["reasons"].([]string)
	if ok {
		return value
	}
	raw, ok := metadata["reasons"].([]any)
	if !ok {
		return nil
	}
	reasons := make([]string, 0, len(raw))
	for _, item := range raw {
		reason := strings.TrimSpace(fmt.Sprint(item))
		if reason != "" {
			reasons = append(reasons, reason)
		}
	}
	return reasons
}

func unsupportedAnswer(answer string) bool {
	return strings.Contains(strings.ToLower(answer), "could not find this in the indexed context")
}

func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
