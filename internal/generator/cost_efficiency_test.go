package generator

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/llm"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

type stubProvider struct {
	requests []llm.CompletionRequest
}

func (p *stubProvider) Name() string { return "anthropic" }

func (p *stubProvider) Configure(config llm.ProviderConfig) error { return nil }

func (p *stubProvider) CountTokens(text string) int { return maxInt(len(text)/4, 1) }

func (p *stubProvider) GetUsage() *llm.UsageMetrics { return &llm.UsageMetrics{} }

func (p *stubProvider) BatchComplete(ctx context.Context, reqs []llm.CompletionRequest) ([]*llm.CompletionResponse, error) {
	responses := make([]*llm.CompletionResponse, 0, len(reqs))
	for _, req := range reqs {
		resp, err := p.Complete(ctx, req)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func (p *stubProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	p.requests = append(p.requests, req)
	if strings.Contains(req.Prompt, "TARGET 1") && strings.Contains(req.Prompt, "TARGET 2") {
		return &llm.CompletionResponse{
			Content:          `{"tests":[{"id":"1","code":"def test_add():\n    assert add(1, 2) == 3"},{"id":"2","code":"def test_subtract():\n    assert subtract(2, 1) == 1"}]}`,
			TokensInput:      400,
			TokensOutput:     200,
			Provider:         "anthropic",
			Model:            llm.AnthropicDefaultModel,
			EstimatedCostUSD: llm.EstimateCost("anthropic", llm.AnthropicDefaultModel, 400, 200),
		}, nil
	}
	return &llm.CompletionResponse{
		Content:          "```python\ndef test_single():\n    assert True\n```",
		TokensInput:      200,
		TokensOutput:     100,
		Provider:         "anthropic",
		Model:            llm.AnthropicDefaultModel,
		EstimatedCostUSD: llm.EstimateCost("anthropic", llm.AnthropicDefaultModel, 200, 100),
	}, nil
	}
}

func TestEngineGenerateArtifactBatchesAndCachesDefinitions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sample.py")
	content := "def add(a, b):\n    return a + b\n\ndef subtract(a, b):\n    return a - b\n"
	if err := os.WriteFile(sourcePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	provider := &stubProvider{}
	engine := &Engine{
		config: EngineConfig{
			Provider:  "anthropic",
			TestTypes: []string{"unit"},
			BatchSize: 2,
		},
		provider: provider,
		cache:    llm.NewCache(100),
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		usage: llm.UsageMetrics{
			Provider: "anthropic",
			Model:    llm.AnthropicDefaultModel,
		},
	}

	adapter := adapters.NewPythonAdapter()
	sourceFile := &models.SourceFile{
		Path:     sourcePath,
		Language: "python",
	}

	first, err := engine.GenerateArtifact(sourceFile, adapter)
	if err != nil {
		t.Fatalf("first GenerateArtifact: %v", err)
	}
	if first.TestCount != 2 {
		t.Fatalf("expected 2 generated tests, got %d", first.TestCount)
	}
	if got := len(provider.requests); got != 1 {
		t.Fatalf("expected one batched provider request, got %d", got)
	}

	usage := engine.GetUsage()
	if usage.BatchCount != 1 || usage.ChunkCount != 1 {
		t.Fatalf("expected one batch/chunk, got batches=%d chunks=%d", usage.BatchCount, usage.ChunkCount)
	}
	if usage.TotalRequests != 1 {
		t.Fatalf("expected one live provider request, got %d", usage.TotalRequests)
	}

	second, err := engine.GenerateArtifact(sourceFile, adapter)
	if err != nil {
		t.Fatalf("second GenerateArtifact: %v", err)
	}
	if second.TestCount != 2 {
		t.Fatalf("expected 2 cached generated tests, got %d", second.TestCount)
	}
	if got := len(provider.requests); got != 1 {
		t.Fatalf("expected cache hit to avoid extra provider calls, got %d requests", got)
	}

	usage = engine.GetUsage()
	if usage.CacheHits != 2 {
		t.Fatalf("expected two cache hits after second pass, got %d", usage.CacheHits)
	}
	if usage.CachedTokens == 0 {
		t.Fatal("expected cached token accounting to be populated")
	}
}

func TestParseChunkResponseRequiresAllTaskIDs(t *testing.T) {
	t.Parallel()

	tasks := []generationTask{
		{id: "1", def: &models.Definition{Name: "add"}},
		{id: "2", def: &models.Definition{Name: "subtract"}},
	}

	_, err := parseChunkResponse(`{"tests":[{"id":"1","code":"def test_add(): pass"}]}`, "python", tasks)
	if err == nil {
		t.Fatal("expected parseChunkResponse to reject missing task ids")
	}
	if !strings.Contains(err.Error(), "missing test code") {
		t.Fatalf("unexpected parse error: %v", err)
	}
}
