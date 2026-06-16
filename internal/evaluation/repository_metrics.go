package evaluation

type LineRange struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type RepositoryQuestionResult struct {
	Question            string            `json:"question"`
	ExpectedFiles       []string          `json:"expected_files"`
	RetrievedFiles      []string          `json:"retrieved_files"`
	ExpectedSymbols     []string          `json:"expected_symbols"`
	RetrievedSymbols    []string          `json:"retrieved_symbols"`
	ExpectedLineRanges  []LineRange       `json:"expected_line_ranges"`
	CitationLineRanges  []LineRange       `json:"citation_line_ranges"`
	ShouldRefuse        bool              `json:"should_refuse"`
	Refused             bool              `json:"refused"`
	SupportedClaimCount int               `json:"supported_claim_count"`
	TotalClaimCount     int               `json:"total_claim_count"`
	LatencyMs           int64             `json:"latency_ms"`
	CostUSD             float64           `json:"cost_usd"`
	StageContributions  map[string]int    `json:"stage_contributions,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

type RepositoryMetrics struct {
	QuestionCount     int     `json:"question_count"`
	FileHitRate       float64 `json:"file_hit_rate"`
	FileCoverage      float64 `json:"file_coverage"`
	SymbolCoverage    float64 `json:"symbol_coverage"`
	LineRangeAccuracy float64 `json:"line_range_accuracy"`
	CitationCoverage  float64 `json:"citation_coverage"`
	EvidenceCoverage  float64 `json:"evidence_coverage"`
	RefusalAccuracy   float64 `json:"refusal_accuracy"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	AvgCostUSD        float64 `json:"avg_cost_usd"`
}

type QuestionComparison struct {
	Question string  `json:"question"`
	Status   string  `json:"status"`
	Delta    float64 `json:"delta"`
}

type RepositoryMetricDelta struct {
	FileHitRate       float64 `json:"file_hit_rate"`
	FileCoverage      float64 `json:"file_coverage"`
	SymbolCoverage    float64 `json:"symbol_coverage"`
	LineRangeAccuracy float64 `json:"line_range_accuracy"`
	CitationCoverage  float64 `json:"citation_coverage"`
	EvidenceCoverage  float64 `json:"evidence_coverage"`
	RefusalAccuracy   float64 `json:"refusal_accuracy"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	AvgCostUSD        float64 `json:"avg_cost_usd"`
}

type RepositoryComparisonSummary struct {
	QuestionsImproved      []QuestionComparison  `json:"questions_improved"`
	QuestionsUnchanged     []QuestionComparison  `json:"questions_unchanged"`
	QuestionsDegraded      []QuestionComparison  `json:"questions_degraded"`
	MetricDelta            RepositoryMetricDelta `json:"metric_delta"`
	StageContributionDelta map[string]int        `json:"stage_contribution_delta"`
}

func ComputeRepositoryMetrics(results []RepositoryQuestionResult) RepositoryMetrics {
	if len(results) == 0 {
		return RepositoryMetrics{}
	}
	var fileHits, citationCovered float64
	var fileCoverage, symbolCoverage, lineAccuracy, evidenceCoverage float64
	var refusalTotal, refusalCorrect float64
	var latency, cost float64
	for _, result := range results {
		fileCoverageForQuestion := coverage(result.ExpectedFiles, result.RetrievedFiles)
		fileCoverage += fileCoverageForQuestion
		if len(result.ExpectedFiles) > 0 && fileCoverageForQuestion > 0 {
			fileHits++
		}
		symbolCoverage += coverage(result.ExpectedSymbols, result.RetrievedSymbols)
		lineAccuracy += lineRangeCoverage(result.ExpectedLineRanges, result.CitationLineRanges)
		if len(result.CitationLineRanges) > 0 || len(result.RetrievedFiles) > 0 {
			citationCovered++
		}
		evidenceCoverage += claimCoverage(result.SupportedClaimCount, result.TotalClaimCount)
		if result.ShouldRefuse {
			refusalTotal++
			if result.Refused {
				refusalCorrect++
			}
		}
		latency += float64(result.LatencyMs)
		cost += result.CostUSD
	}
	n := float64(len(results))
	return RepositoryMetrics{
		QuestionCount:     len(results),
		FileHitRate:       fileHits / n,
		FileCoverage:      fileCoverage / n,
		SymbolCoverage:    symbolCoverage / n,
		LineRangeAccuracy: lineAccuracy / n,
		CitationCoverage:  citationCovered / n,
		EvidenceCoverage:  evidenceCoverage / n,
		RefusalAccuracy:   ratio(refusalCorrect, refusalTotal),
		AvgLatencyMs:      latency / n,
		AvgCostUSD:        cost / n,
	}
}

func CompareRepositoryRuns(baseline, candidate []RepositoryQuestionResult) []QuestionComparison {
	byQuestion := map[string]RepositoryQuestionResult{}
	for _, result := range baseline {
		byQuestion[result.Question] = result
	}
	var comparisons []QuestionComparison
	for _, next := range candidate {
		prev, ok := byQuestion[next.Question]
		if !ok {
			continue
		}
		delta := questionScore(next) - questionScore(prev)
		status := "unchanged"
		if delta > 0.001 {
			status = "improved"
		}
		if delta < -0.001 {
			status = "degraded"
		}
		comparisons = append(comparisons, QuestionComparison{Question: next.Question, Status: status, Delta: delta})
	}
	return comparisons
}

func SummarizeRepositoryComparison(baseline, candidate []RepositoryQuestionResult) RepositoryComparisonSummary {
	comparisons := CompareRepositoryRuns(baseline, candidate)
	summary := RepositoryComparisonSummary{StageContributionDelta: stageContributionDelta(baseline, candidate)}
	for _, comparison := range comparisons {
		switch comparison.Status {
		case "improved":
			summary.QuestionsImproved = append(summary.QuestionsImproved, comparison)
		case "degraded":
			summary.QuestionsDegraded = append(summary.QuestionsDegraded, comparison)
		default:
			summary.QuestionsUnchanged = append(summary.QuestionsUnchanged, comparison)
		}
	}
	before := ComputeRepositoryMetrics(baseline)
	after := ComputeRepositoryMetrics(candidate)
	summary.MetricDelta = RepositoryMetricDelta{
		FileHitRate:       after.FileHitRate - before.FileHitRate,
		FileCoverage:      after.FileCoverage - before.FileCoverage,
		SymbolCoverage:    after.SymbolCoverage - before.SymbolCoverage,
		LineRangeAccuracy: after.LineRangeAccuracy - before.LineRangeAccuracy,
		CitationCoverage:  after.CitationCoverage - before.CitationCoverage,
		EvidenceCoverage:  after.EvidenceCoverage - before.EvidenceCoverage,
		RefusalAccuracy:   after.RefusalAccuracy - before.RefusalAccuracy,
		AvgLatencyMs:      after.AvgLatencyMs - before.AvgLatencyMs,
		AvgCostUSD:        after.AvgCostUSD - before.AvgCostUSD,
	}
	return summary
}

func questionScore(result RepositoryQuestionResult) float64 {
	if result.ShouldRefuse {
		if result.Refused {
			return 1
		}
		return 0
	}
	return (coverage(result.ExpectedFiles, result.RetrievedFiles) +
		coverage(result.ExpectedSymbols, result.RetrievedSymbols) +
		lineRangeCoverage(result.ExpectedLineRanges, result.CitationLineRanges) +
		claimCoverage(result.SupportedClaimCount, result.TotalClaimCount)) / 4
}

func coverage(expected, actual []string) float64 {
	if len(expected) == 0 {
		return 1
	}
	actualSet := set(actual)
	var found float64
	for _, value := range expected {
		if actualSet[value] {
			found++
		}
	}
	return found / float64(len(expected))
}

func lineRangeCoverage(expected, actual []LineRange) float64 {
	if len(expected) == 0 {
		return 1
	}
	var found float64
	for _, want := range expected {
		for _, got := range actual {
			if overlaps(want, got) {
				found++
				break
			}
		}
	}
	return found / float64(len(expected))
}

func overlaps(a, b LineRange) bool {
	if a.Path != b.Path {
		return false
	}
	return a.StartLine <= b.EndLine && b.StartLine <= a.EndLine
}

func claimCoverage(supported, total int) float64 {
	if total <= 0 {
		return 1
	}
	return ratio(float64(supported), float64(total))
}

func ratio(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func stageContributionDelta(baseline, candidate []RepositoryQuestionResult) map[string]int {
	delta := map[string]int{}
	for _, result := range candidate {
		for stage, count := range result.StageContributions {
			delta[stage] += count
		}
	}
	for _, result := range baseline {
		for stage, count := range result.StageContributions {
			delta[stage] -= count
		}
	}
	return delta
}
