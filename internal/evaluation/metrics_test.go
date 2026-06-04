package evaluation

import "testing"

func TestComputeMetrics(t *testing.T) {
	metrics := Compute([]QuestionResult{
		{ExpectedVectorIDs: []string{"a"}, RetrievedVectorIDs: []string{"x", "a"}},
		{ExpectedVectorIDs: []string{"b"}, RetrievedVectorIDs: []string{"z"}},
	})
	if metrics.HitRate != 0.5 {
		t.Fatalf("hit rate = %v", metrics.HitRate)
	}
	if metrics.MRR != 0.25 {
		t.Fatalf("mrr = %v", metrics.MRR)
	}
}
