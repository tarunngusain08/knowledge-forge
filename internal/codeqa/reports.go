package codeqa

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tarunngusain08/knowledge-forge/internal/rag"
)

const maxDeepDiveTargetedRetrievals = 4

type DeepDiveReportResponse struct {
	Summary         string                  `json:"summary"`
	Sections        []DeepDiveReportSection `json:"sections"`
	EvidenceQuality DeepDiveEvidenceQuality `json:"evidence_quality"`
	TraceIDs        []uuid.UUID             `json:"trace_ids"`
	Provenance      AnswerProvenance        `json:"provenance"`
	Model           string                  `json:"model"`
	GeneratedAt     time.Time               `json:"generated_at"`
	Markdown        string                  `json:"markdown"`
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
	FilesExamined    int                `json:"files_examined"`
	CitedFiles       []string           `json:"cited_files"`
	CitedSymbols     []string           `json:"cited_symbols"`
	CitationCount    int                `json:"citation_count"`
	EvidenceCoverage float64            `json:"evidence_coverage"`
	MissingContext   []string           `json:"missing_context"`
	Confidence       EvidenceConfidence `json:"confidence"`
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

	specs := deepDiveReportSectionSpecs()
	sectionAsks := []reportSectionAsk{{
		spec:     specs[0],
		ask:      shared,
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
			spec:        spec,
			ask:         shared,
			targeted:    false,
			missingOnly: true,
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
	sections = append(sections, buildDeepDiveEvidenceQualitySection(shared, quality))
	response := DeepDiveReportResponse{
		Summary:         deepDiveSummary(shared, quality),
		Sections:        sections,
		EvidenceQuality: quality,
		TraceIDs:        deepDiveTraceIDs(sections),
		Provenance:      shared.Provenance,
		Model:           shared.Model,
		GeneratedAt:     generatedAt,
	}
	response.Markdown = formatDeepDiveMarkdown(response)
	return response
}

func buildDeepDiveReportSection(input reportSectionAsk) DeepDiveReportSection {
	missing := deepDiveSectionMissingContext(input)
	findings := deepDiveFindings(input.ask, input.missingOnly)
	return DeepDiveReportSection{
		ID:             input.spec.ID,
		Title:          input.spec.Title,
		Findings:       findings,
		MissingContext: missing,
		Evidence:       evidenceItems(input.ask.Citations),
		Citations:      input.ask.Citations,
		Confidence:     confidenceFromEvidence(input.ask, missing),
		TraceID:        input.ask.TraceID,
		Targeted:       input.targeted,
		Answer:         strings.TrimSpace(input.ask.Answer),
	}
}

func deepDiveSectionMissingContext(input reportSectionAsk) []string {
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

func deepDiveFindings(ask AskResponse, missingOnly bool) []string {
	answer := strings.TrimSpace(ask.Answer)
	if missingOnly || answer == "" || unsupportedAnswer(answer) || len(ask.Citations) == 0 {
		return nil
	}
	return []string{answer}
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
		Findings:       missing,
		MissingContext: missing,
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
		CitedFiles:       files,
		CitedSymbols:     symbols,
		CitationCount:    len(citations),
		EvidenceCoverage: round2(coverage),
		MissingContext:   missing,
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
		Findings:       findings,
		MissingContext: quality.MissingContext,
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
	return citations
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
	return ids
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
