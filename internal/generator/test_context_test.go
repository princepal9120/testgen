package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/llm"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

func TestDiscoverExistingTestContextReadsCompanionTestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	source := filepath.Join(dir, "calculator.py")
	testFile := filepath.Join(dir, "test_calculator.py")
	if err := os.WriteFile(source, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("import pytest\n\ndef test_existing_style():\n    assert True\n"), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}

	ctx := discoverExistingTestContext(source, adapters.NewPythonAdapter(), "")
	if len(ctx.Paths) == 0 || ctx.Paths[0] != testFile {
		t.Fatalf("expected companion test file in context, got %#v", ctx.Paths)
	}
	if !strings.Contains(ctx.Snippet, "test_existing_style") {
		t.Fatalf("expected snippet to include existing test content, got %q", ctx.Snippet)
	}
	if ctx.Framework != "pytest" {
		t.Fatalf("expected pytest framework, got %q", ctx.Framework)
	}
}

func TestDiscoverExistingTestContextReadsMirroredTestsDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	source := filepath.Join(dir, "internal", "calculator", "calculator.go")
	testFile := filepath.Join(dir, "tests", "internal", "calculator", "calculator_test.go")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(testFile), 0o755); err != nil {
		t.Fatalf("mkdir test: %v", err)
	}
	if err := os.WriteFile(source, []byte("package calculator\n\nfunc Add(a, b int) int { return a + b }\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("package calculator\n\nfunc TestExistingStyle(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}

	ctx := discoverExistingTestContext(source, adapters.NewGoAdapter(), "")
	if len(ctx.Paths) == 0 || ctx.Paths[0] != testFile {
		t.Fatalf("expected mirrored test file in context, got %#v", ctx.Paths)
	}
	if !strings.Contains(ctx.Snippet, "TestExistingStyle") {
		t.Fatalf("expected snippet to include mirrored test content, got %q", ctx.Snippet)
	}
}

func TestAugmentPromptWithTestContextAddsProductionStyleInstructions(t *testing.T) {
	t.Parallel()

	prompt := augmentPromptWithTestContext("Generate tests", "vitest", existingTestContext{
		Paths:   []string{"src/foo.test.ts"},
		Snippet: "import { describe, it, expect } from 'vitest'",
	})
	for _, want := range []string{"Match the project's existing test framework", "vitest", "src/foo.test.ts", "describe, it, expect", "Do not invent unavailable helpers"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestBuildGenerationTasksIncludesExistingStyleInPromptAndCacheKey(t *testing.T) {
	t.Parallel()

	provider := llm.NewAnthropicProvider()
	adapter := adapters.NewPythonAdapter()
	defs := []*models.Definition{{Name: "add", Body: "def add(a, b):\n    return a + b"}}

	withoutStyle := buildGenerationTasks(provider, EngineConfig{Provider: "anthropic", TestTypes: []string{"unit"}}, adapter, defs, "", existingTestContext{})
	withStyle := buildGenerationTasks(provider, EngineConfig{Provider: "anthropic", TestTypes: []string{"unit"}}, adapter, defs, "", existingTestContext{
		Framework: "pytest",
		Paths:     []string{"test_calculator.py"},
		Snippet:   "def test_existing_style():\n    assert add(1, 2) == 3",
	})

	if len(withoutStyle) != 1 || len(withStyle) != 1 {
		t.Fatalf("expected one task from each build")
	}
	if withoutStyle[0].cacheKey == withStyle[0].cacheKey {
		t.Fatal("expected style context to affect cache key")
	}
	if !strings.Contains(withStyle[0].prompt, "test_existing_style") {
		t.Fatalf("expected existing test style in prompt, got:\n%s", withStyle[0].prompt)
	}
}
