package evaluation

type QuestionResult struct {
	Question           string   `json:"question"`
	ExpectedVectorIDs  []string `json:"expected_vector_ids"`
	RetrievedVectorIDs []string `json:"retrieved_vector_ids"`
	LatencyMs          int64    `json:"latency_ms"`
}

type Metrics struct {
	QuestionCount int     `json:"question_count"`
	HitRate       float64 `json:"hit_rate"`
	RecallAtK     float64 `json:"recall_at_k"`
	MRR           float64 `json:"mrr"`
}

func Compute(results []QuestionResult) Metrics {
	if len(results) == 0 {
		return Metrics{}
	}
	var hits, reciprocal float64
	var recallSum float64
	for _, result := range results {
		expected := set(result.ExpectedVectorIDs)
		if len(expected) == 0 {
			continue
		}
		found := 0
		firstRank := 0
		for i, id := range result.RetrievedVectorIDs {
			if expected[id] {
				found++
				if firstRank == 0 {
					firstRank = i + 1
				}
			}
		}
		if found > 0 {
			hits++
			reciprocal += 1 / float64(firstRank)
		}
		recallSum += float64(found) / float64(len(expected))
	}
	n := float64(len(results))
	return Metrics{
		QuestionCount: len(results),
		HitRate:       hits / n,
		RecallAtK:     recallSum / n,
		MRR:           reciprocal / n,
	}
}

func set(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}
