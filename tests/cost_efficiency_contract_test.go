package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeCostEstimateJSONIncludesPerFileTokenEstimates(t *testing.T) {
	root := repoRoot(t)
	stdout, stderr, err := runCmdInDir(t, root, "analyze", "--path=tests/testdata/cost/bulk", "--cost-estimate", "--detail=per-file", "--output-format=json")
	if err != nil {
		t.Fatalf("analyze returned error: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}

	files, ok := payload["files"].([]interface{})
	if !ok || len(files) == 0 {
		t.Fatalf("expected per-file details, got %#v", payload["files"])
	}

	for _, raw := range files {
		filePayload, ok := raw.(map[string]interface{})
		if !ok {
			t.Fatalf("expected file payload map, got %#v", raw)
		}
		value, ok := filePayload["estimated_tokens"]
		if !ok {
			t.Fatalf("expected file payload to include estimated_tokens, got %#v", filePayload)
		}
		if tokens := int(value.(float64)); tokens <= 0 {
			t.Fatalf("expected positive estimated_tokens, got %d in %#v", tokens, filePayload)
		}
	}
}

func TestGenerateReportUsageJSONIncludesUsageBlockWithoutAPIKey(t *testing.T) {
	dir := t.TempDir()
	sampleFile := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(sampleFile, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}

	stdout, stderr, err := runCmdInDir(t, dir, "generate", "--file=sample.py", "--dry-run", "--report-usage", "--output-format=json")
	if err != nil {
		t.Fatalf("generate returned error: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}

	usage, ok := payload["usage"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected usage block in generate response, got %#v", payload["usage"])
	}

	for _, key := range []string{"request_count", "cache_hits", "cached_tokens", "estimated_cost_usd"} {
		if _, ok := usage[key]; !ok {
			t.Fatalf("expected usage block to include %q, got %#v", key, usage)
		}
	}
}
