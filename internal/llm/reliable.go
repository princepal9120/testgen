package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ReliableProvider wraps a Provider with retry/backoff behavior for transient failures.
type ReliableProvider struct {
	provider    Provider
	maxAttempts int
	baseDelay   time.Duration
}

// NewReliableProvider wraps a provider with conservative retry behavior.
func NewReliableProvider(provider Provider) Provider {
	return &ReliableProvider{
		provider:    provider,
		maxAttempts: 3,
		baseDelay:   250 * time.Millisecond,
	}
}

func (p *ReliableProvider) Name() string {
	return p.provider.Name()
}

func (p *ReliableProvider) Configure(config ProviderConfig) error {
	return p.provider.Configure(config)
}

func (p *ReliableProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= p.maxAttempts; attempt++ {
		resp, err := p.provider.Complete(ctx, req)
		if err == nil {
			if resp == nil || strings.TrimSpace(resp.Content) == "" {
				lastErr = fmt.Errorf("empty response from provider")
			} else {
				return resp, nil
			}
		} else {
			lastErr = err
			if !isRetryableError(err) || attempt == p.maxAttempts {
				return nil, err
			}
		}

		if attempt < p.maxAttempts {
			delay := time.Duration(attempt) * p.baseDelay
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, lastErr
}

func (p *ReliableProvider) BatchComplete(ctx context.Context, reqs []CompletionRequest) ([]*CompletionResponse, error) {
	return p.provider.BatchComplete(ctx, reqs)
}

func (p *ReliableProvider) CountTokens(text string) int {
	return p.provider.CountTokens(text)
}

func (p *ReliableProvider) GetUsage() *UsageMetrics {
	return p.provider.GetUsage()
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrRateLimited) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "temporar") ||
		strings.Contains(lower, "connection reset") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "status 429") ||
		strings.Contains(lower, "status 500") ||
		strings.Contains(lower, "status 502") ||
		strings.Contains(lower, "status 503") ||
		strings.Contains(lower, "status 504") ||
		strings.Contains(lower, "empty response")
}
