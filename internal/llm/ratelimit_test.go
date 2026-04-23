package llm

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type batchProviderSpy struct {
	mu        sync.Mutex
	calls     int
	batches   [][]CompletionRequest
	responses []*CompletionResponse
}

func (p *batchProviderSpy) Name() string                          { return "batch-spy" }
func (p *batchProviderSpy) Configure(config ProviderConfig) error { return nil }
func (p *batchProviderSpy) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	return nil, errors.New("not implemented")
}
func (p *batchProviderSpy) BatchComplete(ctx context.Context, reqs []CompletionRequest) ([]*CompletionResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	copied := append([]CompletionRequest(nil), reqs...)
	p.batches = append(p.batches, copied)
	return append([]*CompletionResponse(nil), p.responses...), nil
}
func (p *batchProviderSpy) CountTokens(text string) int { return len(text) }
func (p *batchProviderSpy) GetUsage() *UsageMetrics     { return &UsageMetrics{} }

func TestBatcherFlushesWhenBatchSizeReached(t *testing.T) {
	t.Parallel()

	provider := &batchProviderSpy{responses: []*CompletionResponse{{Content: "one"}, {Content: "two"}}}
	batcher := NewBatcher(provider, 2, time.Second)

	batcher.Add(CompletionRequest{Prompt: "first"})
	if pending := batcher.PendingCount(); pending != 1 {
		t.Fatalf("expected 1 pending request, got %d", pending)
	}

	batcher.Add(CompletionRequest{Prompt: "second"})
	if pending := batcher.PendingCount(); pending != 0 {
		t.Fatalf("expected flush to empty pending queue, got %d", pending)
	}
	if provider.calls != 1 {
		t.Fatalf("expected 1 batch call, got %d", provider.calls)
	}
	if len(provider.batches) != 1 || len(provider.batches[0]) != 2 {
		t.Fatalf("unexpected batch capture: %#v", provider.batches)
	}
	if provider.batches[0][0].Prompt != "first" || provider.batches[0][1].Prompt != "second" {
		t.Fatalf("expected request order to be preserved, got %#v", provider.batches[0])
	}
}

func TestBatcherFlushReturnsProviderResponses(t *testing.T) {
	t.Parallel()

	provider := &batchProviderSpy{responses: []*CompletionResponse{{Content: "alpha"}, {Content: "beta"}}}
	batcher := NewBatcher(provider, 3, time.Second)
	batcher.Add(CompletionRequest{Prompt: "alpha"})
	batcher.Add(CompletionRequest{Prompt: "beta"})

	responses, err := batcher.Flush(context.Background())
	if err != nil {
		t.Fatalf("flush returned error: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].Content != "alpha" || responses[1].Content != "beta" {
		t.Fatalf("unexpected responses: %#v", responses)
	}
	if provider.calls != 1 {
		t.Fatalf("expected provider BatchComplete to be called once, got %d", provider.calls)
	}
}

func TestBatcherFlushWithNoPendingRequests(t *testing.T) {
	t.Parallel()

	batcher := NewBatcher(&batchProviderSpy{}, 2, time.Second)
	responses, err := batcher.Flush(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if responses != nil {
		t.Fatalf("expected nil responses for empty flush, got %#v", responses)
	}
}

func TestRateLimiterWaitHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	limiter := NewRateLimiter(1)
	if err := limiter.Wait(context.Background()); err != nil {
		t.Fatalf("initial wait unexpectedly failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := limiter.Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
