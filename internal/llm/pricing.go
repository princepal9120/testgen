package llm

import "strings"

type pricing struct {
	inputPerMillion  float64
	outputPerMillion float64
}

// ResolveProvider normalizes provider names across runtime and analysis flows.
func ResolveProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// ResolveModel returns a concrete model for a provider, falling back to defaults.
func ResolveModel(provider, model string) string {
	model = strings.TrimSpace(model)
	if model != "" {
		return model
	}
	return GetDefaultModel(ResolveProvider(provider))
}

func resolvePricing(provider, model string) pricing {
	provider = ResolveProvider(provider)
	model = ResolveModel(provider, model)

	switch provider {
	case "openai":
		return pricing{inputPerMillion: 10.00, outputPerMillion: 30.00}
	case "gemini":
		switch model {
		case "gemini-1.5-flash", "gemini-1.5-flash-latest":
			return pricing{inputPerMillion: 0.075, outputPerMillion: 0.30}
		default:
			return pricing{inputPerMillion: 1.25, outputPerMillion: 5.00}
		}
	case "groq":
		switch model {
		case "llama-3.1-8b-instant":
			return pricing{inputPerMillion: 0.05, outputPerMillion: 0.08}
		case "mixtral-8x7b-32768":
			return pricing{inputPerMillion: 0.24, outputPerMillion: 0.24}
		default:
			return pricing{inputPerMillion: 0.59, outputPerMillion: 0.79}
		}
	case "anthropic":
		fallthrough
	default:
		return pricing{inputPerMillion: 3.00, outputPerMillion: 15.00}
	}
}

// EstimateCost returns the provider-aware cost estimate for the supplied tokens.
func EstimateCost(provider, model string, inputTokens, outputTokens int) float64 {
	rates := resolvePricing(provider, model)
	return (float64(inputTokens) * rates.inputPerMillion / 1_000_000) +
		(float64(outputTokens) * rates.outputPerMillion / 1_000_000)
}
