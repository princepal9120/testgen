package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/princepal9120/testgen-cli/pkg/models"
)

func TestServiceGenerateReturnsNoSourceFilesError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	service := NewService()

	_, err := service.Generate(context.Background(), GenerateRequest{
		Path:      dir,
		Recursive: true,
		TestTypes: []string{"unit"},
	})
	if err == nil {
		t.Fatal("expected error when no source files are present")
	}
}

func TestServiceGenerateScansAndReturnsResultsForDefinitionFreeFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(file, []byte("# no functions here\n"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}

	service := NewService()
	resp, err := service.Generate(context.Background(), GenerateRequest{
		File:      file,
		TestTypes: []string{"unit"},
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("generate returned error: %v", err)
	}
	if len(resp.SourceFiles) != 1 {
		t.Fatalf("expected 1 source file, got %d", len(resp.SourceFiles))
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Error != nil {
		t.Fatalf("expected no generation error, got %v", resp.Results[0].Error)
	}
}

func TestServiceAnalyzeReturnsStructuredStats(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.py"), []byte("def main():\n    return 42\n"), 0o644); err != nil {
		t.Fatalf("write python file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "util.js"), []byte("function add(a, b) {\n  return a + b\n}\n"), 0o644); err != nil {
		t.Fatalf("write js file: %v", err)
	}

	service := NewService()
	resp, err := service.Analyze(context.Background(), AnalyzeRequest{
		Path:         dir,
		Recursive:    true,
		CostEstimate: true,
		Detail:       "per-file",
	})
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if resp.TotalFiles != 2 {
		t.Fatalf("expected 2 files, got %d", resp.TotalFiles)
	}
	if len(resp.ByLanguage) != 2 {
		t.Fatalf("expected stats for 2 languages, got %d", len(resp.ByLanguage))
	}
	if len(resp.Files) != 2 {
		t.Fatalf("expected 2 file entries, got %d", len(resp.Files))
	}
	if resp.EstimatedTokens == 0 {
		t.Fatal("expected cost estimation to populate tokens")
	}
	if resp.ExactFunctionFiles != 2 {
		t.Fatalf("expected exact function counts for both files, got %d", resp.ExactFunctionFiles)
	}
	if resp.HeuristicFunctionFiles != 0 {
		t.Fatalf("expected no heuristic fallback files, got %d", resp.HeuristicFunctionFiles)
	}
	for _, file := range resp.Files {
		if file.FunctionCountMode != "exact" {
			t.Fatalf("expected exact function count mode, got %q for %s", file.FunctionCountMode, file.Path)
		}
	}
}

func TestServiceValidateFindsExistingTests(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	testsDir := filepath.Join(dir, "tests")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.MkdirAll(testsDir, 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}

	sourcePath := filepath.Join(srcDir, "math.py")
	testPath := filepath.Join(testsDir, "test_math.py")
	if err := os.WriteFile(sourcePath, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(testPath, []byte("def test_add():\n    assert True\n"), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}

	service := NewService()
	resp, err := service.Validate(context.Background(), ValidateRequest{
		Path:      srcDir,
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
	if resp.Result.FilesWithTests != 1 {
		t.Fatalf("expected 1 file with tests, got %d", resp.Result.FilesWithTests)
	}
	if len(resp.Result.FilesMissingTests) != 0 {
		t.Fatalf("expected no missing tests, got %d", len(resp.Result.FilesMissingTests))
	}
}

func TestPatchFromResultBuildsStructuredPatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := &models.GenerationResult{
		SourceFile: &models.SourceFile{
			Path:     filepath.Join(dir, "math.py"),
			Language: "python",
		},
		TestPath: filepath.Join(dir, "tests", "test_math.py"),
		TestCode: "def test_add():\n    assert True\n",
	}

	patch := patchFromResult(result)
	if patch == nil {
		t.Fatal("expected patch to be generated")
	}
	if patch.Action != "create_or_replace" {
		t.Fatalf("expected create_or_replace action, got %q", patch.Action)
	}
	if patch.Path != result.TestPath {
		t.Fatalf("expected patch path %q, got %q", result.TestPath, patch.Path)
	}
}

func TestArtifactFromResultIncludesErrorDetails(t *testing.T) {
	t.Parallel()

	result := &models.GenerationResult{
		SourceFile: &models.SourceFile{
			Path:     "sample.py",
			Language: "python",
		},
		TestPath:        "tests/test_sample.py",
		TestCode:        "def test_sample(): pass",
		FunctionsTested: []string{"sample"},
		Error:           errors.New("validation failed: boom"),
	}

	artifact := artifactFromResult(result)
	if !artifact.Generated {
		t.Fatal("expected artifact to be marked as generated")
	}
	if !artifact.ValidationFailed {
		t.Fatal("expected validation failure to be detected")
	}
	if artifact.Error == "" {
		t.Fatal("expected artifact error message")
	}
}
