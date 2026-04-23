package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/llm"
)

type stubProvider struct {
	requests []llm.CompletionRequest
}

func (p *stubProvider) Name() string                          { return "anthropic" }
func (p *stubProvider) Configure(config llm.ProviderConfig) error { return nil }
func (p *stubProvider) CountTokens(text string) int           { return maxInt(len(text)/4, 1) }
func (p *stubProvider) GetUsage() *llm.UsageMetrics           { return &llm.UsageMetrics{} }
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

func TestEngineGenerateArtifactBatchesAndCachesDefinitions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sample.py")
	content := "def add(a, b):\n    return a + b\n\ndef subtract(a, b):\n    return a - b\n"
	if err := os.WriteFile(sourcePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	engine := &Engine{
		config: EngineConfig{
			Provider:  "anthropic",
			TestTypes: []string{"unit"},
			BatchSize: 2,
		},
		provider: &stubProvider{},
		cache:    llm.NewCache(100),
	}

	adapter := adapters.NewPythonAdapter()
	sourceFile := &adapters.SourceFileToModelsHack
	_ = sourceFile
}
