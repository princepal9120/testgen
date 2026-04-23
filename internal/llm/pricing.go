package llm

import "strings"

const (
	defaultEstimatedPromptTokensPerFunction = 150
	defaultEstimatedOutputTokensPerFunction = 200
	defaultEstimatedSystemPromptTokens      = 500
	defaultEstimatedBatchSize               = 5
)

// UsageEstimate captures provider-aware offline usage estimates.
type UsageEstimate struct {
	Provider                string
	Model                   string
	Requests                int
	BatchCount              int
	ChunkCount              int
	InputTokens             int
	OutputTokens            int
	TotalTokens             int
	EstimatedCostUSD        float64
	InputCostPerMillionUSD  float64
	OutputCostPerMillionUSD float64
}

// EstimateOfflineUsage computes a provider-aware offline estimate for a generation workload.
func EstimateOfflineUsage(provider string, model string, functionCount int, batchSize int) UsageEstimate {
	provider = ResolveProvider(provider)
	model = ResolveModel(provider, model)
	if batchSize <= 0 {
		batchSize = defaultEstimatedBatchSize
	}

	inputRate, outputRate := pricingForProviderModel(provider, model)
	if functionCount <= 0 {
		return UsageEstimate{
			Provider:                provider,
			Model:                   model,
			InputCostPerMillionUSD:  inputRate,
			OutputCostPerMillionUSD: outputRate,
		}
	}

	batchCount := ceilDiv(functionCount, batchSize)
	inputTokens := (functionCount * defaultEstimatedPromptTokensPerFunction) + (batchCount * defaultEstimatedSystemPromptTokens)
	outputTokens := functionCount * defaultEstimatedOutputTokensPerFunction
	totalTokens := inputTokens + outputTokens

	return UsageEstimate{
		Provider:                provider,
		Model:                   model,
		Requests:                batchCount,
		BatchCount:              batchCount,
		ChunkCount:              batchCount,
		InputTokens:             inputTokens,
		OutputTokens:            outputTokens,
		TotalTokens:             totalTokens,
		EstimatedCostUSD:        (float64(inputTokens) * inputRate / 1_000_000) + (float64(outputTokens) * outputRate / 1_000_000),
		InputCostPerMillionUSD:  inputRate,
		OutputCostPerMillionUSD: outputRate,
	}
}

// ResolveProvider normalizes a provider identifier onto the supported provider set.
func ResolveProvider(provider string) string {
	return normalizeProvider(provider)
}

// ResolveModel returns the requested model or the provider default when omitted.
func ResolveModel(provider string, model string) string {
	model = strings.TrimSpace(model)
	if model != "" {
		return model
	}
	return GetDefaultModel(ResolveProvider(provider))
}

// EstimateCost computes provider/model-aware cost for concrete input/output token counts.
func EstimateCost(provider string, model string, inputTokens int, outputTokens int) float64 {
	provider = ResolveProvider(provider)
	model = ResolveModel(provider, model)
	inputRate, outputRate := pricingForProviderModel(provider, model)
	return (float64(inputTokens) * inputRate / 1_000_000) + (float64(outputTokens) * outputRate / 1_000_000)
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return "openai"
	case "gemini":
		return "gemini"
	case "groq":
		return "groq"
	default:
		return "anthropic"
	}
}

func pricingForProviderModel(provider string, model string) (float64, float64) {
	normalizedModel := strings.ToLower(strings.TrimSpace(model))

	switch normalizeProvider(provider) {
	case "openai":
		return 10.00, 30.00
	case "gemini":
		if strings.Contains(normalizedModel, "flash") {
			return 0.075, 0.30
		}
		return 1.25, 5.00
	case "groq":
		switch {
		case strings.Contains(normalizedModel, "llama-3.1-8b-instant"):
			return 0.05, 0.08
		case strings.Contains(normalizedModel, "mixtral-8x7b-32768"):
			return 0.24, 0.24
		default:
			return 0.59, 0.79
		}
	default:
		return 3.00, 15.00
	}
}

func ceilDiv(n, d int) int {
	if n <= 0 || d <= 0 {
		return 0
	}
	return (n + d - 1) / d
}
