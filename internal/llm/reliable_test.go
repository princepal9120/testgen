package llm

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	attempts int
	errs     []error
	resp     *CompletionResponse
}

func (p *fakeProvider) Name() string                          { return "fake" }
func (p *fakeProvider) Configure(config ProviderConfig) error { return nil }
func (p *fakeProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	p.attempts++
	if len(p.errs) >= p.attempts && p.errs[p.attempts-1] != nil {
		return nil, p.errs[p.attempts-1]
	}
	return p.resp, nil
}
func (p *fakeProvider) BatchComplete(ctx context.Context, reqs []CompletionRequest) ([]*CompletionResponse, error) {
	return nil, nil
}
func (p *fakeProvider) CountTokens(text string) int { return len(text) }
func (p *fakeProvider) GetUsage() *UsageMetrics     { return &UsageMetrics{} }

func TestReliableProviderRetriesRateLimit(t *testing.T) {
	t.Parallel()

	base := &fakeProvider{
		errs: []error{ErrRateLimited, nil},
		resp: &CompletionResponse{Content: "ok"},
	}
	provider := NewReliableProvider(base)

	resp, err := provider.Complete(context.Background(), CompletionRequest{Prompt: "hi"})
	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if resp.Content != "ok" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if base.attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", base.attempts)
	}
}

func TestReliableProviderDoesNotRetryPermanentError(t *testing.T) {
	t.Parallel()

	base := &fakeProvider{
		errs: []error{errors.New("invalid request")},
	}
	provider := NewReliableProvider(base)

	_, err := provider.Complete(context.Background(), CompletionRequest{Prompt: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
	if base.attempts != 1 {
		t.Fatalf("expected single attempt, got %d", base.attempts)
	}
}
