package llm

import "testing"

func TestEstimateCostUsesProviderPricing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		model    string
		input    int
		output   int
		want     float64
	}{
		{name: "anthropic", provider: "anthropic", model: AnthropicDefaultModel, input: 1_000_000, output: 1_000_000, want: 18.00},
		{name: "openai", provider: "openai", model: OpenAIDefaultModel, input: 1_000_000, output: 1_000_000, want: 40.00},
		{name: "gemini flash", provider: "gemini", model: "gemini-1.5-flash", input: 1_000_000, output: 1_000_000, want: 0.375},
		{name: "gemini pro", provider: "gemini", model: "gemini-1.5-pro", input: 1_000_000, output: 1_000_000, want: 6.25},
		{name: "groq llama70b", provider: "groq", model: "llama-3.3-70b-versatile", input: 1_000_000, output: 1_000_000, want: 1.38},
		{name: "groq mixtral", provider: "groq", model: "mixtral-8x7b-32768", input: 1_000_000, output: 1_000_000, want: 0.48},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := EstimateCost(tt.provider, tt.model, tt.input, tt.output); got != tt.want {
				t.Fatalf("EstimateCost(%q, %q) = %v, want %v", tt.provider, tt.model, got, tt.want)
			}
		})
	}
}
