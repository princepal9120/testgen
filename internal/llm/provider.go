/*
Package llm provides LLM provider abstraction for test generation.

This package implements a provider interface supporting multiple LLM backends
(Anthropic Claude, OpenAI GPT) with cost optimization features like caching
and batching.
*/
package llm

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrNoAPIKey      = errors.New("API key not configured")
	ErrRateLimited   = errors.New("rate limited by provider")
	ErrContextLength = errors.New("context length exceeded")
	ErrInvalidModel  = errors.New("invalid model specified")
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider name (e.g., "anthropic", "openai")
	Name() string

	// Configure sets up the provider with credentials
	Configure(config ProviderConfig) error

	// Complete sends a prompt and returns a completion
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// BatchComplete processes multiple prompts
	BatchComplete(ctx context.Context, reqs []CompletionRequest) ([]*CompletionResponse, error)

	// CountTokens estimates token count for text
	CountTokens(text string) int

	// GetUsage returns usage metrics
	GetUsage() *UsageMetrics
}

// ProviderConfig contains provider configuration
type ProviderConfig struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float32
	BaseURL     string // Optional custom endpoint
}

// CompletionRequest represents a completion request
type CompletionRequest struct {
	Prompt      string
	SystemRole  string
	MaxTokens   int
	Temperature float32
	Seed        *int // For reproducibility
}

// CompletionResponse represents a completion response
type CompletionResponse struct {
	Content          string  `json:"content"`
	TokensInput      int     `json:"tokens_input,omitempty"`
	TokensOutput     int     `json:"tokens_output,omitempty"`
	Cached           bool    `json:"cached,omitempty"`
	Provider         string  `json:"provider,omitempty"`
	Model            string  `json:"model,omitempty"`
	FinishReason     string  `json:"finish_reason,omitempty"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd,omitempty"`
}

// UsageMetrics tracks API usage
type UsageMetrics struct {
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	Estimated        bool    `json:"estimated"`
	TotalRequests    int     `json:"request_count"`
	BatchCount       int     `json:"batch_count"`
	ChunkCount       int     `json:"chunk_count"`
	CacheHits        int     `json:"cache_hits"`
	CacheMisses      int     `json:"cache_misses"`
	TotalTokensIn    int     `json:"total_tokens_in"`
	TotalTokensOut   int     `json:"total_tokens_out"`
	CachedTokens     int     `json:"cached_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// Clone returns a defensive copy of usage metrics.
func (u UsageMetrics) Clone() *UsageMetrics {
	copy := u
	return &copy
}

// TotalTokens returns the aggregate input and output token count.
func (u UsageMetrics) TotalTokens() int {
	return u.TotalTokensIn + u.TotalTokensOut
}

// CacheHitRate returns the derived cache hit rate.
func (u UsageMetrics) CacheHitRate() float64 {
	total := u.CacheHits + u.CacheMisses
	if total == 0 {
		return 0
	}
	return float64(u.CacheHits) / float64(total)
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DefaultModels for each provider
const (
	AnthropicDefaultModel = "claude-3-5-sonnet-20241022"
	OpenAIDefaultModel    = "gpt-4-turbo-preview"
	GeminiDefaultModel    = "gemini-1.5-pro"
	GroqDefaultModel      = "llama-3.3-70b-versatile"
)

// GetDefaultModel returns the default model for a provider
func GetDefaultModel(providerName string) string {
	switch providerName {
	case "anthropic":
		return AnthropicDefaultModel
	case "openai":
		return OpenAIDefaultModel
	case "gemini":
		return GeminiDefaultModel
	case "groq":
		return GroqDefaultModel
	default:
		return ""
	}
}

// NewProvider returns a concrete provider implementation for the requested provider name.
func NewProvider(providerName string) Provider {
	switch ResolveProvider(providerName) {
	case "openai":
		return NewOpenAIProvider()
	case "gemini":
		return NewGeminiProvider()
	case "groq":
		return NewGroqProvider()
	default:
		return NewAnthropicProvider()
	}
}
