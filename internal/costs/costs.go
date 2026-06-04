package costs

type Usage struct {
	Provider     string
	Model        string
	Operation    string
	InputTokens  int
	OutputTokens int
}

type Pricing struct {
	InputPerMillionUSD  float64
	OutputPerMillionUSD float64
}

var defaultPricing = map[string]Pricing{
	"vertex:gemini-2.5-flash:generate": {
		InputPerMillionUSD:  0.30,
		OutputPerMillionUSD: 2.50,
	},
	"vertex:gemini-embedding-001:embed": {
		InputPerMillionUSD: 0.15,
	},
	"vertex:semantic-ranker-default@latest:rerank": {
		InputPerMillionUSD: 0.00,
	},
}

func EstimateUSD(usage Usage) float64 {
	key := usage.Provider + ":" + usage.Model + ":" + usage.Operation
	pricing, ok := defaultPricing[key]
	if !ok {
		return 0
	}
	inputCost := float64(usage.InputTokens) / 1_000_000 * pricing.InputPerMillionUSD
	outputCost := float64(usage.OutputTokens) / 1_000_000 * pricing.OutputPerMillionUSD
	return inputCost + outputCost
}
