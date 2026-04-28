package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/adapters"
)

const maxTestContextBytes = 6000

type existingTestContext struct {
	Framework string
	Paths     []string
	Snippet   string
}

func discoverExistingTestContext(sourcePath string, adapter adapters.LanguageAdapter, configuredFramework string) existingTestContext {
	ctx := existingTestContext{Framework: strings.TrimSpace(configuredFramework)}
	if sourcePath == "" || adapter == nil {
		return ctx
	}

	if ctx.Framework == "" {
		ctx.Framework = adapter.SelectFramework(findProjectRoot(filepath.Dir(sourcePath)))
	}

	candidates := candidateTestFiles(sourcePath, adapter)
	seen := map[string]bool{}
	var snippets []string
	remaining := maxTestContextBytes

	for _, path := range candidates {
		clean := filepath.Clean(path)
		if seen[clean] || remaining <= 0 {
			continue
		}
		seen[clean] = true

		content, err := os.ReadFile(clean)
		if err != nil || len(strings.TrimSpace(string(content))) == 0 {
			continue
		}

		ctx.Paths = append(ctx.Paths, clean)
		chunk := truncateForPrompt(string(content), remaining)
		snippets = append(snippets, fmt.Sprintf("Existing test file: %s\n%s", clean, chunk))
		remaining -= len(chunk)
	}

	ctx.Snippet = strings.TrimSpace(strings.Join(snippets, "\n\n---\n\n"))
	return ctx
}

func candidateTestFiles(sourcePath string, adapter adapters.LanguageAdapter) []string {
	dir := filepath.Dir(sourcePath)
	projectRoot := findProjectRoot(dir)
	relSource, relErr := filepath.Rel(projectRoot, sourcePath)
	expected := adapter.GenerateTestPath(sourcePath, "")
	candidates := []string{expected}

	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			if isLikelyTestFile(path, adapter.GetLanguage()) {
				candidates = append(candidates, path)
			}
		}
	}

	testsDir := filepath.Join(dir, "__tests__")
	if entries, err := os.ReadDir(testsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			path := filepath.Join(testsDir, entry.Name())
			if isLikelyTestFile(path, adapter.GetLanguage()) {
				candidates = append(candidates, path)
			}
		}
	}

	// Common mirrored test layouts, for example:
	//   src/foo/bar.py -> tests/foo/test_bar.py
	//   internal/pkg/foo.go -> tests/internal/pkg/foo_test.go
	// Keep this narrow and deterministic so prompt context stays useful.
	if relErr == nil && relSource != "" && !strings.HasPrefix(relSource, "..") {
		for _, rootTestsDir := range []string{"tests", "test", "spec", "__tests__"} {
			mirrored := filepath.Join(projectRoot, rootTestsDir, relSource)
			candidates = append(candidates, mirroredTestPathVariants(mirrored, adapter.GetLanguage())...)
		}
	}

	// Include a few top-level test files when the source has no direct companion.
	// This helps new modules inherit project conventions without dragging the whole
	// test suite into the prompt.
	for _, rootTestsDir := range []string{"tests", "test", "spec", "__tests__"} {
		entries, err := os.ReadDir(filepath.Join(projectRoot, rootTestsDir))
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			path := filepath.Join(projectRoot, rootTestsDir, entry.Name())
			if isLikelyTestFile(path, adapter.GetLanguage()) {
				candidates = append(candidates, path)
			}
		}
	}

	return uniqueSorted(candidates)
}

func mirroredTestPathVariants(path, language string) []string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	switch language {
	case "go":
		return []string{filepath.Join(dir, stem+"_test"+ext)}
	case "python":
		return []string{filepath.Join(dir, "test_"+base), filepath.Join(dir, stem+"_test"+ext)}
	case "javascript", "typescript":
		return []string{filepath.Join(dir, stem+".test"+ext), filepath.Join(dir, stem+".spec"+ext)}
	case "rust":
		return []string{filepath.Join(dir, stem+"_test"+ext), filepath.Join(dir, base)}
	case "java":
		return []string{filepath.Join(dir, stem+"Test"+ext)}
	default:
		return []string{path}
	}
}

func isLikelyTestFile(path, language string) bool {
	base := strings.ToLower(filepath.Base(path))
	dir := strings.ToLower(filepath.Base(filepath.Dir(path)))
	switch language {
	case "go":
		return strings.HasSuffix(base, "_test.go")
	case "python":
		return strings.HasSuffix(base, "_test.py") || strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py")
	case "javascript", "typescript":
		return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") || dir == "__tests__"
	case "rust":
		return strings.HasSuffix(base, "_test.rs") || dir == "tests"
	case "java":
		return strings.HasSuffix(base, "test.java") || strings.Contains(filepath.ToSlash(strings.ToLower(path)), "/src/test/")
	default:
		return false
	}
}

func findProjectRoot(start string) string {
	current := filepath.Clean(start)
	markers := []string{"go.mod", "package.json", "pyproject.toml", "pytest.ini", "Cargo.toml", "pom.xml", "build.gradle", "settings.gradle", ".git"}
	for {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			return start
		}
		current = parent
	}
}

func truncateForPrompt(content string, limit int) string {
	content = strings.TrimSpace(content)
	if limit <= 0 || len(content) <= limit {
		return content
	}
	if limit < 200 {
		return content[:limit]
	}
	return strings.TrimSpace(content[:limit-120]) + "\n... [truncated existing test context]"
}

func uniqueSorted(paths []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		clean := filepath.Clean(path)
		if clean == "." || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	sort.Strings(out)
	return out
}
