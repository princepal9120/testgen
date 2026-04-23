package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/metrics"
)

func TestServiceAnalyzePersistsCostMetrics(t *testing.T) {
	fixturePath, err := filepath.Abs(filepath.Join("..", "..", "tests", "testdata", "cost", "bulk"))
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}

	workingDir := t.TempDir()
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWD)
	}()

	service := NewService()
	resp, err := service.Analyze(context.Background(), AnalyzeRequest{
		Path:         fixturePath,
		Recursive:    true,
		CostEstimate: true,
		Detail:       "per-file",
	})
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if resp.EstimatedTokens == 0 {
		t.Fatal("expected estimated tokens to be populated")
	}

	entries, err := os.ReadDir(filepath.Join(workingDir, ".testgen", "metrics"))
	if err != nil {
		t.Fatalf("read metrics dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 metrics file, got %d", len(entries))
	}

	data, err := os.ReadFile(filepath.Join(workingDir, ".testgen", "metrics", entries[0].Name()))
	if err != nil {
		t.Fatalf("read metrics file: %v", err)
	}

	var saved metrics.RunMetrics
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}

	if saved.Operation != "analyze" {
		t.Fatalf("expected analyze operation, got %q", saved.Operation)
	}
	if saved.TargetPath != resp.Path {
		t.Fatalf("expected target path %q, got %q", resp.Path, saved.TargetPath)
	}
	if saved.TotalFiles != resp.TotalFiles {
		t.Fatalf("expected total files %d, got %d", resp.TotalFiles, saved.TotalFiles)
	}
	if saved.TokensInput != resp.EstimatedTokens {
		t.Fatalf("expected tokens input %d, got %d", resp.EstimatedTokens, saved.TokensInput)
	}
	if saved.TotalCostUSD != resp.EstimatedCost {
		t.Fatalf("expected estimated cost %f, got %f", resp.EstimatedCost, saved.TotalCostUSD)
	}
	if saved.ExactFunctionFiles != resp.ExactFunctionFiles {
		t.Fatalf("expected exact function files %d, got %d", resp.ExactFunctionFiles, saved.ExactFunctionFiles)
	}
	if saved.SuccessCount != resp.TotalFiles {
		t.Fatalf("expected success count %d, got %d", resp.TotalFiles, saved.SuccessCount)
	}
}
