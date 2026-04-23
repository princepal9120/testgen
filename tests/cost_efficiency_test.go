package tests

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func TestAnalyzeCostEstimateJSONFixture(t *testing.T) {
	root := repoRoot(t)
	stdout, stderr, err := runCmdInDir(t, root, "analyze", "--path=tests/testdata/cost/bulk", "--cost-estimate", "--detail=per-file", "--output-format=json")
	if err != nil {
		t.Fatalf("analyze returned error: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if strings.Contains(stderr, "Usage:") {
		t.Fatalf("expected JSON mode to suppress usage banners, stderr=%s", stderr)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}
	if payload["success"] != true {
		t.Fatalf("expected success=true, got %#v", payload["success"])
	}
	if got := int(payload["total_files"].(float64)); got != 3 {
		t.Fatalf("expected 3 total files, got %d", got)
	}
	if got := int(payload["estimated_tokens"].(float64)); got <= 0 {
		t.Fatalf("expected positive estimated tokens, got %d", got)
	}
	files, ok := payload["files"].([]interface{})
	if !ok || len(files) != 3 {
		t.Fatalf("expected 3 file entries, got %#v", payload["files"])
	}
}

func TestAnalyzeCostEstimateTextFixture(t *testing.T) {
	root := repoRoot(t)
	stdout, stderr, err := runCmdInDir(t, root, "analyze", "--path=tests/testdata/cost/repeated", "--cost-estimate", "--detail=per-file")
	if err != nil {
		t.Fatalf("analyze returned error: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
	}

	combined := stdout + stderr
	for _, needle := range []string{"=== Codebase Analysis ===", "--- Cost Estimate ---", "--- Per-File Details ---", "service.py", "helpers.py"} {
		if !strings.Contains(combined, needle) {
			t.Fatalf("expected analyze output to contain %q, got:\n%s", needle, combined)
		}
	}
}
