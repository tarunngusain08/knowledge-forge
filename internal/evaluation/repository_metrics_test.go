package evaluation

import "testing"

func TestComputeRepositoryMetrics(t *testing.T) {
	metrics := ComputeRepositoryMetrics([]RepositoryQuestionResult{
		{
			Question:            "Where is auth implemented?",
			ExpectedFiles:       []string{"internal/auth/service.go", "cmd/api/main.go"},
			RetrievedFiles:      []string{"internal/auth/service.go"},
			ExpectedSymbols:     []string{"AuthService"},
			RetrievedSymbols:    []string{"AuthService"},
			ExpectedLineRanges:  []LineRange{{Path: "internal/auth/service.go", StartLine: 10, EndLine: 30}},
			CitationLineRanges:  []LineRange{{Path: "internal/auth/service.go", StartLine: 20, EndLine: 25}},
			SupportedClaimCount: 3,
			TotalClaimCount:     4,
			LatencyMs:           100,
			CostUSD:             0.01,
		},
		{
			Question:     "What handles payroll?",
			ShouldRefuse: true,
			Refused:      true,
			LatencyMs:    50,
			CostUSD:      0.005,
		},
	})

	if metrics.QuestionCount != 2 {
		t.Fatalf("question count = %d", metrics.QuestionCount)
	}
	if metrics.FileHitRate != 0.5 {
		t.Fatalf("file hit rate = %v", metrics.FileHitRate)
	}
	if metrics.FileCoverage != 0.75 {
		t.Fatalf("file coverage = %v", metrics.FileCoverage)
	}
	if metrics.SymbolCoverage != 1 {
		t.Fatalf("symbol coverage = %v", metrics.SymbolCoverage)
	}
	if metrics.LineRangeAccuracy != 1 {
		t.Fatalf("line accuracy = %v", metrics.LineRangeAccuracy)
	}
	if metrics.RefusalAccuracy != 1 {
		t.Fatalf("refusal accuracy = %v", metrics.RefusalAccuracy)
	}
}

func TestCompareRepositoryRuns(t *testing.T) {
	comparisons := CompareRepositoryRuns(
		[]RepositoryQuestionResult{{
			Question:       "Where is auth implemented?",
			ExpectedFiles:  []string{"auth.go"},
			RetrievedFiles: []string{"main.go"},
		}},
		[]RepositoryQuestionResult{{
			Question:       "Where is auth implemented?",
			ExpectedFiles:  []string{"auth.go"},
			RetrievedFiles: []string{"auth.go"},
		}},
	)
	if len(comparisons) != 1 {
		t.Fatalf("comparisons = %d", len(comparisons))
	}
	if comparisons[0].Status != "improved" {
		t.Fatalf("status = %s", comparisons[0].Status)
	}
}

func TestSummarizeRepositoryComparison(t *testing.T) {
	summary := SummarizeRepositoryComparison(
		[]RepositoryQuestionResult{{
			Question:            "Where is auth implemented?",
			ExpectedFiles:       []string{"auth.go"},
			RetrievedFiles:      []string{"main.go"},
			SupportedClaimCount: 1,
			TotalClaimCount:     2,
			LatencyMs:           100,
			CostUSD:             0.01,
			StageContributions:  map[string]int{"dense": 1},
		}},
		[]RepositoryQuestionResult{{
			Question:            "Where is auth implemented?",
			ExpectedFiles:       []string{"auth.go"},
			RetrievedFiles:      []string{"auth.go"},
			SupportedClaimCount: 2,
			TotalClaimCount:     2,
			LatencyMs:           140,
			CostUSD:             0.012,
			StageContributions:  map[string]int{"dense": 1, "symbol": 1},
		}},
	)

	if len(summary.QuestionsImproved) != 1 {
		t.Fatalf("improved questions = %d", len(summary.QuestionsImproved))
	}
	if summary.MetricDelta.FileCoverage <= 0 {
		t.Fatalf("file coverage delta = %v", summary.MetricDelta.FileCoverage)
	}
	if summary.MetricDelta.AvgLatencyMs != 40 {
		t.Fatalf("latency delta = %v", summary.MetricDelta.AvgLatencyMs)
	}
	if summary.StageContributionDelta["symbol"] != 1 {
		t.Fatalf("symbol stage delta = %v", summary.StageContributionDelta["symbol"])
	}
}
