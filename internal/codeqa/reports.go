package codeqa

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/codeintel"
	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

const maxDeepDiveTargetedRetrievals = 4

type DeepDiveReportResponse struct {
	Summary                string                  `json:"summary"`
	Sections               []DeepDiveReportSection `json:"sections"`
	EvidenceQuality        DeepDiveEvidenceQuality `json:"evidence_quality"`
	TraceIDs               []uuid.UUID             `json:"trace_ids"`
	Provenance             AnswerProvenance        `json:"provenance"`
	Model                  string                  `json:"model"`
	GeneratedAt            time.Time               `json:"generated_at"`
	Markdown               string                  `json:"markdown"`
	ClaimGroundingMappings []ClaimGroundingMapping `json:"claim_grounding_mappings"`
	ClaimGroundingCoverage float64                 `json:"claim_grounding_coverage"`
}

type DeepDiveReportSection struct {
	ID             string             `json:"id"`
	Title          string             `json:"title"`
	Findings       []string           `json:"findings"`
	MissingContext []string           `json:"missing_context"`
	Evidence       []EvidenceItem     `json:"evidence"`
	Citations      []rag.Citation     `json:"citations"`
	Confidence     EvidenceConfidence `json:"confidence"`
	TraceID        uuid.UUID          `json:"trace_id,omitempty"`
	Targeted       bool               `json:"targeted"`
	Answer         string             `json:"answer,omitempty"`
}

type DeepDiveEvidenceQuality struct {
	FilesExamined          int                     `json:"files_examined"`
	CitedFiles             []string                `json:"cited_files"`
	CitedSymbols           []string                `json:"cited_symbols"`
	CitationCount          int                     `json:"citation_count"`
	EvidenceCoverage       float64                 `json:"evidence_coverage"`
	ClaimGroundingCoverage float64                 `json:"claim_grounding_coverage"`
	ClaimGroundingMappings []ClaimGroundingMapping `json:"claim_grounding_mappings"`
	MissingContext         []string                `json:"missing_context"`
	Confidence             EvidenceConfidence      `json:"confidence"`
}

type ClaimGroundingMapping struct {
	Claim      string `json:"claim"`
	CitationID string `json:"citation_id"`
	File       string `json:"file"`
	LineRange  string `json:"line_range"`
	Evidence   string `json:"evidence"`
}

type architectureChunkStore interface {
	ListArchitectureChunks(ctx context.Context, repositoryID, snapshotID uuid.UUID) ([]codeintel.Chunk, error)
}

type reportSectionSpec struct {
	ID           string
	Title        string
	Prompt       string
	TargetedByV1 bool
}

type reportSectionAsk struct {
	spec        reportSectionSpec
	ask         AskResponse
	targeted    bool
	missingOnly bool
}

func (s *Service) GenerateDeepDiveReport(ctx context.Context, req WorkflowRequest) (DeepDiveReportResponse, error) {
	shared, err := s.Ask(ctx, AskRequest{
		UserID:          req.UserID,
		RepositoryID:    req.RepositoryID,
		BranchName:      req.BranchName,
		Question:        deepDiveSharedPrompt(req.Request),
		TopK:            workflowTopK(req.TopK),
		RerankerEnabled: req.RerankerEnabled,
	})
	if err != nil {
		return DeepDiveReportResponse{}, err
	}
	architectureAsk := shared
	architectureCitations, err := s.architectureEvidenceCitations(ctx, shared.Provenance)
	if err != nil {
		return DeepDiveReportResponse{}, err
	}
	if len(architectureCitations) > 0 {
		architectureAsk.Citations = mergeCitations(architectureAsk.Citations, architectureCitations)
	}

	specs := deepDiveReportSectionSpecs()
	sectionAsks := []reportSectionAsk{{
		spec:     specs[0],
		ask:      architectureAsk,
		targeted: false,
	}}
	targetedCount := 0
	for _, spec := range specs[1:] {
		if spec.TargetedByV1 && targetedCount < maxDeepDiveTargetedRetrievals {
			targeted, err := s.Ask(ctx, AskRequest{
				UserID:          req.UserID,
				RepositoryID:    req.RepositoryID,
				BranchName:      req.BranchName,
				Question:        deepDiveSectionPrompt(req.Request, spec),
				TopK:            workflowTopK(req.TopK),
				RerankerEnabled: req.RerankerEnabled,
			})
			if err != nil {
				return DeepDiveReportResponse{}, err
			}
			sectionAsks = append(sectionAsks, reportSectionAsk{
				spec:     spec,
				ask:      targeted,
				targeted: true,
			})
			targetedCount++
			continue
		}
		sectionAsks = append(sectionAsks, reportSectionAsk{
			spec:     spec,
			ask:      shared,
			targeted: false,
		})
	}

	return buildDeepDiveReportResponse(shared, sectionAsks, time.Now().UTC()), nil
}

func deepDiveReportSectionSpecs() []reportSectionSpec {
	return []reportSectionSpec{
		{
			ID:     "architecture_overview",
			Title:  "Architecture Overview",
			Prompt: "Describe the repository architecture, major subsystems, and how the pieces fit together.",
		},
		{
			ID:           "entry_points",
			Title:        "Entry Points",
			Prompt:       "Identify executable or service entry points and cite the files that start the application.",
			TargetedByV1: true,
		},
		{
			ID:     "main_packages",
			Title:  "Main Packages",
			Prompt: "Identify the main packages or modules and explain their responsibilities from cited evidence.",
		},
		{
			ID:           "authentication_flow",
			Title:        "Authentication Flow",
			Prompt:       "Explain how authentication or user identity is implemented. If there is no evidence, say so.",
			TargetedByV1: true,
		},
		{
			ID:           "data_layer",
			Title:        "Data Layer",
			Prompt:       "Identify database, persistence, migration, or repository-layer code with citations.",
			TargetedByV1: true,
		},
		{
			ID:     "external_services",
			Title:  "External Services",
			Prompt: "Identify external providers, APIs, cloud services, or third-party integrations from cited files.",
		},
		{
			ID:           "testing_strategy",
			Title:        "Testing Strategy",
			Prompt:       "Explain the test strategy and cite representative test files.",
			TargetedByV1: true,
		},
		{
			ID:     "risk_areas",
			Title:  "Risk Areas",
			Prompt: "Identify evidence-backed technical risk areas. Do not invent risks that are not supported by cited files.",
		},
		{
			ID:     "suggested_improvements",
			Title:  "Suggested Improvements",
			Prompt: "Suggest improvements supported by repository evidence and cite why each suggestion is grounded.",
		},
	}
}

func buildDeepDiveReportResponse(shared AskResponse, sectionAsks []reportSectionAsk, generatedAt time.Time) DeepDiveReportResponse {
	sections := make([]DeepDiveReportSection, 0, len(sectionAsks)+2)
	for _, sectionAsk := range sectionAsks {
		sections = append(sections, buildDeepDiveReportSection(sectionAsk))
	}
	missingSection := buildDeepDiveMissingContextSection(shared, sections)
	sections = append(sections, missingSection)
	quality := buildDeepDiveEvidenceQuality(sections)
	mappings := buildClaimGroundingMappings(sections)
	claimCoverage := claimGroundingCoverage(sections, mappings)
	quality.ClaimGroundingMappings = nonNilClaimGroundingMappings(mappings)
	quality.ClaimGroundingCoverage = claimCoverage
	sections = append(sections, buildDeepDiveEvidenceQualitySection(shared, quality))
	response := DeepDiveReportResponse{
		Summary:                deepDiveSummary(shared, quality),
		Sections:               sections,
		EvidenceQuality:        quality,
		TraceIDs:               deepDiveTraceIDs(sections),
		Provenance:             shared.Provenance,
		Model:                  shared.Model,
		GeneratedAt:            generatedAt,
		ClaimGroundingMappings: nonNilClaimGroundingMappings(mappings),
		ClaimGroundingCoverage: claimCoverage,
	}
	response.Markdown = formatDeepDiveMarkdown(response)
	return response
}

func buildDeepDiveReportSection(input reportSectionAsk) DeepDiveReportSection {
	missing := deepDiveSectionMissingContext(input)
	findings := deepDiveFindingsForSection(input)
	return DeepDiveReportSection{
		ID:             input.spec.ID,
		Title:          input.spec.Title,
		Findings:       nonNilStrings(findings),
		MissingContext: nonNilStrings(missing),
		Evidence:       nonNilEvidenceItems(evidenceItems(input.ask.Citations)),
		Citations:      nonNilCitations(input.ask.Citations),
		Confidence:     confidenceFromEvidence(input.ask, missing),
		TraceID:        input.ask.TraceID,
		Targeted:       input.targeted,
		Answer:         strings.TrimSpace(input.ask.Answer),
	}
}

func deepDiveSectionMissingContext(input reportSectionAsk) []string {
	if input.spec.ID == "architecture_overview" && len(architectureFindingsFromCitations(input.ask.Citations)) == 0 && len(input.ask.Citations) > 0 {
		return []string{"Architecture overview did not find enough code-structure evidence; docs may support but cannot define architecture by themselves."}
	}
	if input.missingOnly {
		return []string{"This section did not receive targeted retrieval in v1; the shared evidence pass was insufficient to make section-specific claims."}
	}
	if len(input.ask.Citations) == 0 || unsupportedAnswer(input.ask.Answer) {
		return []string{"No cited repository evidence was retrieved for this report section."}
	}
	missing := []string{}
	if input.ask.Provenance.CommitSHA == "" {
		missing = append(missing, "Repository commit SHA is unavailable, so this section is not fully reproducible.")
	}
	if len(missing) == 0 {
		missing = append(missing, "No missing context was detected from retrieved report evidence.")
	}
	return missing
}

func deepDiveFindingsForSection(input reportSectionAsk) []string {
	if input.spec.ID == "architecture_overview" {
		return architectureFindingsFromCitations(input.ask.Citations)
	}
	return deepDiveFindings(input.ask, input.missingOnly)
}

func deepDiveFindings(ask AskResponse, missingOnly bool) []string {
	answer := strings.TrimSpace(ask.Answer)
	if missingOnly || answer == "" || unsupportedAnswer(answer) || len(ask.Citations) == 0 {
		return []string{}
	}
	return []string{answer}
}

func architectureFindingsFromCitations(citations []rag.Citation) []string {
	layerEvidence := map[string]string{}
	for _, citation := range citations {
		if !sourceCodeArchitectureCitation(citation) {
			continue
		}
		path := strings.TrimSpace(citation.Path)
		if path == "" {
			continue
		}
		label := architectureLayerForPath(path)
		if label == "" || layerEvidence[label] != "" {
			continue
		}
		layerEvidence[label] = path
	}
	order := []string{
		"API layer",
		"retrieval/RAG layer",
		"UI layer",
		"data layer",
		"indexing/worker layer",
		"provider integration layer",
		"evaluation layer",
		"deployment layer",
	}
	findings := make([]string, 0, len(layerEvidence))
	for _, label := range order {
		path := layerEvidence[label]
		if path == "" {
			continue
		}
		findings = append(findings, fmt.Sprintf("Repository structure identifies the %s through `%s`.", label, path))
	}
	return findings
}

func (s *Service) architectureEvidenceCitations(ctx context.Context, provenance AnswerProvenance) ([]rag.Citation, error) {
	store, ok := s.repos.(architectureChunkStore)
	if !ok || provenance.RepositoryID == uuid.Nil || provenance.SnapshotID == uuid.Nil {
		return []rag.Citation{}, nil
	}
	chunks, err := store.ListArchitectureChunks(ctx, provenance.RepositoryID, provenance.SnapshotID)
	if err != nil {
		return nil, err
	}
	citations := make([]rag.Citation, 0, len(chunks))
	for _, chunk := range chunks {
		if !sourceCodeArchitecturePath(chunk.Path) || architectureLayerForPath(chunk.Path) == "" {
			continue
		}
		metadata := chunk.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}
		metadata["path"] = chunk.Path
		metadata["start_line"] = chunk.StartLine
		metadata["end_line"] = chunk.EndLine
		metadata["evidence_groups"] = architectureEvidenceGroupsForPath(chunk.Path)
		citations = append(citations, rag.Citation{
			ChunkID:      chunk.ID,
			DocumentID:   chunk.FileID,
			RepositoryID: chunk.RepositoryID,
			SnapshotID:   chunk.SnapshotID,
			BranchName:   provenance.BranchName,
			CommitSHA:    provenance.CommitSHA,
			Path:         chunk.Path,
			StartLine:    chunk.StartLine,
			EndLine:      chunk.EndLine,
			Excerpt:      excerptForEvidence(chunk.Content),
			Metadata:     metadata,
		})
	}
	return mergeCitations(nil, citations), nil
}

func sourceCodeArchitectureCitation(citation rag.Citation) bool {
	return citation.StartLine > 0 && citation.EndLine >= citation.StartLine && sourceCodeArchitecturePath(citation.Path)
}

func sourceCodeArchitecturePath(path string) bool {
	normalized := strings.ToLower(strings.TrimSpace(path))
	if normalized == "" || strings.HasSuffix(normalized, ".md") || strings.HasSuffix(normalized, ".txt") || strings.HasSuffix(normalized, ".keep") {
		return false
	}
	return strings.HasSuffix(normalized, ".go") ||
		strings.HasSuffix(normalized, ".ts") ||
		strings.HasSuffix(normalized, ".tsx") ||
		strings.HasSuffix(normalized, ".py") ||
		strings.HasSuffix(normalized, ".sql")
}

func architectureEvidenceGroupsForPath(path string) []string {
	switch architectureLayerForPath(path) {
	case "API layer":
		return []string{"api_layer"}
	case "UI layer":
		return []string{"ui_layer"}
	case "retrieval/RAG layer":
		return []string{"retrieval_layer", "rag_context"}
	default:
		return []string{}
	}
}

func architectureLayerForPath(path string) string {
	normalized := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(path)), "./")
	switch {
	case strings.HasPrefix(normalized, "cmd/api/") || strings.HasPrefix(normalized, "internal/http") || strings.HasPrefix(normalized, "internal/api"):
		return "API layer"
	case strings.HasPrefix(normalized, "internal/rag/") || strings.HasPrefix(normalized, "internal/retrieval/") || strings.HasPrefix(normalized, "internal/codeqa/"):
		return "retrieval/RAG layer"
	case strings.HasPrefix(normalized, "ui/web/") || strings.HasPrefix(normalized, "ui/streamlit/"):
		return "UI layer"
	case strings.HasPrefix(normalized, "internal/db/") || strings.HasPrefix(normalized, "internal/repositories/") || strings.HasPrefix(normalized, "migrations/") || strings.HasPrefix(normalized, "queries/"):
		return "data layer"
	case strings.HasPrefix(normalized, "cmd/worker/") || strings.HasPrefix(normalized, "internal/indexing/") || strings.HasPrefix(normalized, "internal/worker"):
		return "indexing/worker layer"
	case strings.HasPrefix(normalized, "internal/providers/"):
		return "provider integration layer"
	case strings.HasPrefix(normalized, "internal/evaluation/") || strings.HasPrefix(normalized, "eval-runner/"):
		return "evaluation layer"
	case strings.HasPrefix(normalized, "deploy/") || strings.HasPrefix(normalized, ".github/") || strings.HasPrefix(normalized, "docker"):
		return "deployment layer"
	default:
		return ""
	}
}

func buildDeepDiveMissingContextSection(shared AskResponse, sections []DeepDiveReportSection) DeepDiveReportSection {
	missing := aggregateDeepDiveMissingContext(sections)
	if len(missing) == 0 {
		missing = []string{"No missing context was detected across generated report sections."}
	}
	ask := shared
	ask.Citations = nil
	return DeepDiveReportSection{
		ID:             "missing_context",
		Title:          "Missing Context",
		Findings:       nonNilStrings(missing),
		MissingContext: nonNilStrings(missing),
		Evidence:       []EvidenceItem{},
		Citations:      []rag.Citation{},
		Confidence: EvidenceConfidence{
			Label:             "Medium",
			Score:             0.5,
			EvidenceCoverage:  0,
			CitationCount:     0,
			ContextTokenCount: shared.Provenance.ContextTokenCount,
			Reasons:           []string{"missing context is derived from section-level evidence checks"},
		},
		TraceID: shared.TraceID,
	}
}

func buildDeepDiveEvidenceQuality(sections []DeepDiveReportSection) DeepDiveEvidenceQuality {
	citations := uniqueDeepDiveCitations(sections)
	files := citationFiles(citations)
	symbols := citationSymbols(citations)
	missing := aggregateDeepDiveMissingContext(sections)
	contentSections := 0
	supportedSections := 0
	contextTokens := 0
	for _, section := range sections {
		if section.ID == "missing_context" || section.ID == "evidence_quality" {
			continue
		}
		contentSections++
		contextTokens += section.Confidence.ContextTokenCount
		if len(section.Citations) > 0 && len(section.Findings) > 0 {
			supportedSections++
		}
	}
	coverage := 0.0
	if contentSections > 0 {
		coverage = float64(supportedSections) / float64(contentSections)
	}
	label := "Low"
	if coverage >= 0.75 && len(citations) >= 5 {
		label = "High"
	} else if coverage >= 0.4 && len(citations) > 0 {
		label = "Medium"
	}
	reasons := []string{
		fmt.Sprintf("%d of %d report sections have cited findings", supportedSections, contentSections),
		fmt.Sprintf("%d unique cited files were examined", len(files)),
	}
	if len(missing) > 0 {
		reasons = append(reasons, "missing-context checks reduce confidence")
	}
	return DeepDiveEvidenceQuality{
		FilesExamined:    len(files),
		CitedFiles:       nonNilStrings(files),
		CitedSymbols:     nonNilStrings(symbols),
		CitationCount:    len(citations),
		EvidenceCoverage: round2(coverage),
		MissingContext:   nonNilStrings(missing),
		Confidence: EvidenceConfidence{
			Label:             label,
			Score:             round2(coverage),
			EvidenceCoverage:  round2(coverage),
			CitationCount:     len(citations),
			ContextTokenCount: contextTokens,
			Reasons:           reasons,
		},
	}
}

func buildDeepDiveEvidenceQualitySection(shared AskResponse, quality DeepDiveEvidenceQuality) DeepDiveReportSection {
	findings := []string{
		fmt.Sprintf("Files examined: %d", quality.FilesExamined),
		fmt.Sprintf("Unique citations: %d", quality.CitationCount),
		fmt.Sprintf("Evidence coverage: %.0f%%", quality.EvidenceCoverage*100),
		fmt.Sprintf("Claim grounding coverage: %.0f%%", quality.ClaimGroundingCoverage*100),
		fmt.Sprintf("Confidence: %s", quality.Confidence.Label),
	}
	if len(quality.CitedFiles) > 0 {
		findings = append(findings, "Cited files: "+strings.Join(quality.CitedFiles, ", "))
	}
	if len(quality.CitedSymbols) > 0 {
		findings = append(findings, "Cited symbols: "+strings.Join(quality.CitedSymbols, ", "))
	}
	return DeepDiveReportSection{
		ID:             "evidence_quality",
		Title:          "Evidence Quality",
		Findings:       nonNilStrings(findings),
		MissingContext: nonNilStrings(quality.MissingContext),
		Evidence:       []EvidenceItem{},
		Citations:      []rag.Citation{},
		Confidence:     quality.Confidence,
		TraceID:        shared.TraceID,
	}
}

func aggregateDeepDiveMissingContext(sections []DeepDiveReportSection) []string {
	seen := map[string]bool{}
	var missing []string
	for _, section := range sections {
		for _, item := range section.MissingContext {
			item = strings.TrimSpace(item)
			if item == "" || strings.HasPrefix(item, "No missing context") || seen[item] {
				continue
			}
			seen[item] = true
			missing = append(missing, item)
		}
	}
	sort.Strings(missing)
	return missing
}

func uniqueDeepDiveCitations(sections []DeepDiveReportSection) []rag.Citation {
	seen := map[string]bool{}
	var citations []rag.Citation
	for _, section := range sections {
		for _, citation := range section.Citations {
			key := citation.ChunkID.String()
			if citation.ChunkID == uuid.Nil {
				key = fmt.Sprintf("%s:%d:%d:%s", citation.Path, citation.StartLine, citation.EndLine, citation.Excerpt)
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			citations = append(citations, citation)
		}
	}
	return nonNilCitations(citations)
}

func mergeCitations(primary, secondary []rag.Citation) []rag.Citation {
	seen := map[string]bool{}
	merged := make([]rag.Citation, 0, len(primary)+len(secondary))
	for _, citation := range append(primary, secondary...) {
		key := deepDiveCitationKey(citation)
		if seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, citation)
	}
	return nonNilCitations(merged)
}

func deepDiveCitationKey(citation rag.Citation) string {
	if citation.ChunkID != uuid.Nil {
		return citation.ChunkID.String()
	}
	return fmt.Sprintf("%s:%d:%d:%s", citation.Path, citation.StartLine, citation.EndLine, citation.Excerpt)
}

func buildClaimGroundingMappings(sections []DeepDiveReportSection) []ClaimGroundingMapping {
	var mappings []ClaimGroundingMapping
	for _, section := range sections {
		if section.ID == "missing_context" || section.ID == "evidence_quality" || len(section.Citations) == 0 {
			continue
		}
		for _, finding := range section.Findings {
			claim := strings.TrimSpace(finding)
			if claim == "" {
				continue
			}
			for _, citation := range section.Citations {
				if citation.Path == "" || citation.StartLine == 0 || citation.EndLine == 0 {
					continue
				}
				mappings = append(mappings, ClaimGroundingMapping{
					Claim:      claim,
					CitationID: citationID(citation),
					File:       citation.Path,
					LineRange:  fmt.Sprintf("%s:%d-%d", citation.Path, citation.StartLine, citation.EndLine),
					Evidence:   excerptForEvidence(citation.Excerpt),
				})
			}
		}
	}
	return nonNilClaimGroundingMappings(mappings)
}

func claimGroundingCoverage(sections []DeepDiveReportSection, mappings []ClaimGroundingMapping) float64 {
	totalClaims := 0
	supported := map[string]bool{}
	for _, mapping := range mappings {
		if mapping.Claim != "" {
			supported[mapping.Claim] = true
		}
	}
	for _, section := range sections {
		if section.ID == "missing_context" || section.ID == "evidence_quality" {
			continue
		}
		for _, finding := range section.Findings {
			if strings.TrimSpace(finding) != "" {
				totalClaims++
			}
		}
	}
	if totalClaims == 0 {
		return 0
	}
	return round2(float64(len(supported)) / float64(totalClaims))
}

func citationID(citation rag.Citation) string {
	if citation.ChunkID != uuid.Nil {
		return citation.ChunkID.String()
	}
	return fmt.Sprintf("%s:%d-%d", citation.Path, citation.StartLine, citation.EndLine)
}

func excerptForEvidence(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 240 {
		return value
	}
	return strings.TrimSpace(value[:240])
}

func nonNilClaimGroundingMappings(values []ClaimGroundingMapping) []ClaimGroundingMapping {
	if values == nil {
		return []ClaimGroundingMapping{}
	}
	return values
}

func deepDiveTraceIDs(sections []DeepDiveReportSection) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	var ids []uuid.UUID
	for _, section := range sections {
		if section.TraceID == uuid.Nil || seen[section.TraceID] {
			continue
		}
		seen[section.TraceID] = true
		ids = append(ids, section.TraceID)
	}
	if ids == nil {
		return []uuid.UUID{}
	}
	return ids
}

func nonNilCitations(values []rag.Citation) []rag.Citation {
	if values == nil {
		return []rag.Citation{}
	}
	return values
}

func nonNilEvidenceItems(values []EvidenceItem) []EvidenceItem {
	if values == nil {
		return []EvidenceItem{}
	}
	return values
}

func deepDiveSummary(shared AskResponse, quality DeepDiveEvidenceQuality) string {
	if len(shared.Citations) == 0 || unsupportedAnswer(shared.Answer) {
		return "Knowledge Forge could not produce a grounded repository deep-dive from the indexed evidence."
	}
	commit := shared.Provenance.CommitSHA
	if commit == "" {
		commit = "unknown commit"
	}
	return fmt.Sprintf("Knowledge Forge generated a cited repository deep-dive for %s with %d unique citations across %d files.",
		commit, quality.CitationCount, quality.FilesExamined)
}

func deepDiveSharedPrompt(request string) string {
	if strings.TrimSpace(request) == "" {
		request = "Generate a repository due-diligence deep-dive report."
	}
	return fmt.Sprintf(`Repository deep-dive request:
%s

Use only retrieved repository context. First find broad evidence about architecture, entry points, main packages, authentication, data layer, external services, tests, risks, and suggested improvements.
Every claim must be supported by cited repository files. If evidence is insufficient, state missing context instead of inventing details.`, strings.TrimSpace(request))
}

func deepDiveSectionPrompt(request string, spec reportSectionSpec) string {
	focus := strings.TrimSpace(request)
	if focus == "" {
		focus = "Generate a repository due-diligence deep-dive report."
	}
	return fmt.Sprintf(`Repository deep-dive section request:
%s

Section: %s
Task: %s

Use only retrieved repository context. Every claim must cite evidence. If evidence is insufficient, state missing context.`, focus, spec.Title, spec.Prompt)
}

func formatDeepDiveMarkdown(report DeepDiveReportResponse) string {
	var builder strings.Builder
	builder.WriteString("# Repository Deep-Dive Report\n\n")
	builder.WriteString(fmt.Sprintf("Generated: %s\n\n", report.GeneratedAt.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("- Repository ID: `%s`\n", report.Provenance.RepositoryID))
	builder.WriteString(fmt.Sprintf("- Branch: `%s`\n", emptyFallback(report.Provenance.BranchName, "unknown")))
	builder.WriteString(fmt.Sprintf("- Snapshot ID: `%s`\n", report.Provenance.SnapshotID))
	builder.WriteString(fmt.Sprintf("- Commit SHA: `%s`\n", emptyFallback(report.Provenance.CommitSHA, "unknown")))
	builder.WriteString(fmt.Sprintf("- Model: `%s`\n\n", emptyFallback(report.Model, "unknown")))
	builder.WriteString("## Summary\n\n")
	builder.WriteString(report.Summary + "\n\n")
	for _, section := range report.Sections {
		builder.WriteString(fmt.Sprintf("## %s\n\n", section.Title))
		if section.Confidence.Label != "" {
			builder.WriteString(fmt.Sprintf("Confidence: **%s**", section.Confidence.Label))
			if section.Confidence.Score > 0 {
				builder.WriteString(fmt.Sprintf(" (%.0f%%)", section.Confidence.Score*100))
			}
			builder.WriteString("\n\n")
		}
		if len(section.Findings) > 0 {
			builder.WriteString("### Findings\n\n")
			writeMarkdownList(&builder, section.Findings)
		}
		if len(section.MissingContext) > 0 {
			builder.WriteString("### Missing Context\n\n")
			writeMarkdownList(&builder, section.MissingContext)
		}
		if len(section.Citations) > 0 {
			builder.WriteString("### Evidence\n\n")
			for _, citation := range section.Citations {
				builder.WriteString(fmt.Sprintf("- `%s:%d-%d`", emptyFallback(citation.Path, "unknown"), citation.StartLine, citation.EndLine))
				if citation.CommitSHA != "" {
					builder.WriteString(fmt.Sprintf(" @ `%s`", citation.CommitSHA))
				}
				if citation.Excerpt != "" {
					builder.WriteString(fmt.Sprintf(" - %s", sanitizeMarkdownLine(citation.Excerpt)))
				}
				builder.WriteString("\n")
			}
			builder.WriteString("\n")
		}
	}
	return strings.TrimSpace(builder.String()) + "\n"
}

func writeMarkdownList(builder *strings.Builder, values []string) {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		builder.WriteString("- " + sanitizeMarkdownLine(value) + "\n")
	}
	builder.WriteString("\n")
}

func sanitizeMarkdownLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
