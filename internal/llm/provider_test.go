package llm

import "testing"

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{name: "anthropic", provider: "anthropic", want: AnthropicDefaultModel},
		{name: "openai", provider: "openai", want: OpenAIDefaultModel},
		{name: "gemini", provider: "gemini", want: GeminiDefaultModel},
		{name: "groq", provider: "groq", want: GroqDefaultModel},
		{name: "unknown", provider: "unknown", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetDefaultModel(tc.provider)
			if got != tc.want {
				t.Fatalf("GetDefaultModel(%q) = %q, want %q", tc.provider, got, tc.want)
			}
		})
	}
}

func TestEstimateOfflineUsageUsesProviderAwarePricing(t *testing.T) {
	t.Parallel()

	geminiFlash := EstimateOfflineUsage("gemini", "gemini-1.5-flash", 4, 2)
	geminiPro := EstimateOfflineUsage("gemini", "gemini-1.5-pro", 4, 2)
	if geminiFlash.Provider != "gemini" {
		t.Fatalf("expected normalized provider, got %q", geminiFlash.Provider)
	}
	if geminiFlash.BatchCount != 2 || geminiFlash.Requests != 2 {
		t.Fatalf("expected 2 batches/requests, got batches=%d requests=%d", geminiFlash.BatchCount, geminiFlash.Requests)
	}
	if geminiFlash.EstimatedCostUSD >= geminiPro.EstimatedCostUSD {
		t.Fatalf("expected flash pricing to be cheaper than pro, flash=%f pro=%f", geminiFlash.EstimatedCostUSD, geminiPro.EstimatedCostUSD)
	}

	groqFast := EstimateOfflineUsage("groq", "llama-3.1-8b-instant", 4, 2)
	groqLarge := EstimateOfflineUsage("groq", "llama-3.3-70b-versatile", 4, 2)
	if groqFast.EstimatedCostUSD >= groqLarge.EstimatedCostUSD {
		t.Fatalf("expected 8b instant pricing to be cheaper than 70b versatile, instant=%f versatile=%f", groqFast.EstimatedCostUSD, groqLarge.EstimatedCostUSD)
	}
}

func TestEstimateOfflineUsageFallsBackToProviderDefaults(t *testing.T) {
	t.Parallel()

	estimate := EstimateOfflineUsage("", "", 3, 0)
	if estimate.Provider != "anthropic" {
		t.Fatalf("expected default provider anthropic, got %q", estimate.Provider)
	}
	if estimate.Model != AnthropicDefaultModel {
		t.Fatalf("expected default model %q, got %q", AnthropicDefaultModel, estimate.Model)
	}
	if estimate.BatchCount != 1 {
		t.Fatalf("expected default batch count 1, got %d", estimate.BatchCount)
	}
	if estimate.TotalTokens != estimate.InputTokens+estimate.OutputTokens {
		t.Fatalf("expected total tokens to equal input+output, got total=%d input=%d output=%d", estimate.TotalTokens, estimate.InputTokens, estimate.OutputTokens)
	}
}
