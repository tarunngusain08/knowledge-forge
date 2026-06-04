package costs

import "testing"

func TestEstimateUSD(t *testing.T) {
	got := EstimateUSD(Usage{
		Provider:     "vertex",
		Model:        "gemini-2.5-flash",
		Operation:    "generate",
		InputTokens:  1_000_000,
		OutputTokens: 1_000_000,
	})
	if got != 2.80 {
		t.Fatalf("expected 2.80, got %.2f", got)
	}
}
