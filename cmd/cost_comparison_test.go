package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/spf13/viper"
)

func TestRunCostOutputsJSONEstimate(t *testing.T) {
	resetCostCommandState()
	viper.Reset()
	logger = nil

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(sourceFile, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	costPath = dir
	costProvider = "gemini"
	costOutputFormat = "json"

	stdout := captureStdout(t, func() error {
		return runCost(costCmd, nil)
	})

	var resp app.AnalyzeResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, stdout)
	}
	if !resp.Success {
		t.Fatalf("expected success response, got %#v", resp)
	}
	if resp.Provider != "gemini" || resp.Model == "" {
		t.Fatalf("expected provider/model metadata, got provider=%q model=%q", resp.Provider, resp.Model)
	}
	if resp.EstimatedCost <= 0 || resp.EstimatedTokens <= 0 || resp.EstimatedRequests <= 0 {
		t.Fatalf("expected populated cost estimate, got cost=%f tokens=%d requests=%d", resp.EstimatedCost, resp.EstimatedTokens, resp.EstimatedRequests)
	}
}

func TestRunComparisonOutputsLLMVsSkillRows(t *testing.T) {
	resetComparisonCommandState()
	viper.Reset()
	logger = nil

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(sourceFile, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	comparisonPath = dir
	comparisonOutputFormat = "json"

	stdout := captureStdout(t, func() error {
		return runComparison(comparisonCmd, nil)
	})

	var resp comparisonResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, stdout)
	}
	if !resp.Success {
		t.Fatalf("expected successful comparison response")
	}
	if len(resp.Rows) == 0 || !strings.Contains(resp.Rows[0].TestGen, "Scans") {
		t.Fatalf("expected comparison rows, got %#v", resp.Rows)
	}
	if len(resp.Commands) != 3 {
		t.Fatalf("expected three recommended commands, got %d", len(resp.Commands))
	}
}

func TestGenerateCommandHasTestcaseAlias(t *testing.T) {
	found := false
	for _, alias := range generateCmd.Aliases {
		if alias == "testcase" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected generate command to expose testcase alias")
	}
}

func resetCostCommandState() {
	costPath = "."
	costRecursive = true
	costDetail = "summary"
	costProvider = ""
	costModel = ""
	costBatchSize = 5
	costOutputFormat = "text"
	quiet = false
	verbose = false
	costCmd.SilenceErrors = false
	costCmd.SilenceUsage = false
}

func resetComparisonCommandState() {
	comparisonPath = "."
	comparisonRecursive = true
	comparisonProvider = ""
	comparisonModel = ""
	comparisonBatchSize = 5
	comparisonOutputFormat = "text"
	quiet = false
	verbose = false
	comparisonCmd.SilenceErrors = false
	comparisonCmd.SilenceUsage = false
}
