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
