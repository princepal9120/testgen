/*
Package validation provides test validation and execution functionality.
*/
package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/scanner"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

// Config contains validation configuration
type Config struct {
	MinCoverage   float64
	FailOnMissing bool
	ReportGaps    bool
}

// Result represents validation results
type Result struct {
	CoveragePercent   float64  `json:"coverage_percent"`
	FilesWithTests    int      `json:"files_with_tests"`
	FilesMissingTests []string `json:"files_missing_tests"`
	TestsPassed       int      `json:"tests_passed"`
	TestsFailed       int      `json:"tests_failed"`
	Errors            []string `json:"errors,omitempty"`
}

// Validator validates tests
type Validator struct {
	config Config
}

type discoveredTest struct {
	path   string
	inline bool
}

// NewValidator creates a new validator
func NewValidator(config Config) *Validator {
	return &Validator{
		config: config,
	}
}

// Validate validates tests for the given source files.
func (v *Validator) Validate(path string, sourceFiles []*models.SourceFile) (*Result, error) {
	result := &Result{
		FilesMissingTests: make([]string, 0),
		Errors:            make([]string, 0),
	}

	if len(sourceFiles) == 0 {
		return result, nil
	}

	parser := NewCoverageParser()
	coveredLanguages := make(map[string]bool)
	languageRoots := make(map[string]string)
	coverageSamples := 0
	totalCoverage := 0.0

	for _, sf := range sourceFiles {
		if sf == nil {
			continue
		}

		discovered := discoverTestsForSource(sf)
		validTests := 0
		for _, test := range discovered {
			if test.inline {
				validTests++
				continue
			}

			looksLikeTest, err := looksLikeTestFile(sf.Language, test.path)
			if err != nil {
				result.Errors = append(result.Errors, err.Error())
				continue
			}
			if !looksLikeTest {
				result.Errors = append(result.Errors, fmt.Sprintf("%s: discovered test file does not contain recognizable %s tests", test.path, scanner.NormalizeLanguage(sf.Language)))
				continue
			}

			validTests++
		}

		if validTests == 0 {
			result.FilesMissingTests = append(result.FilesMissingTests, sf.Path)
			continue
		}

		result.FilesWithTests++
		language := scanner.NormalizeLanguage(sf.Language)
		coveredLanguages[language] = true
		languageRoots[language] = mergeExecutionRoot(languageRoots[language], executionRoot(path, sf.Path, discovered))
	}

	for _, language := range sortedKeys(coveredLanguages) {
		adapter := adapters.DefaultRegistry().GetAdapter(language)
		if adapter == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("no adapter registered for %s validation", language))
			continue
		}

		testRoot := languageRoots[language]
		if testRoot == "" {
			testRoot = path
		}

		testResults, err := adapter.RunTests(testRoot)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s tests: %v", language, err))
			continue
		}

		result.TestsPassed += testResults.PassedCount
		result.TestsFailed += testResults.FailedCount
		result.Errors = append(result.Errors, testResults.Errors...)

		coverage := testResults.Coverage
		if coverage == 0 && testResults.Output != "" {
			coverage = parser.ParseCoverage(testResults.Output, language)
		}
		if coverage > 0 {
			totalCoverage += coverage
			coverageSamples++
		}
	}

	if coverageSamples > 0 {
		result.CoveragePercent = totalCoverage / float64(coverageSamples)
	} else {
		result.CoveragePercent = float64(result.FilesWithTests) / float64(len(sourceFiles)) * 100
	}

	if v.config.FailOnMissing && len(result.FilesMissingTests) > 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("%d file(s) are missing tests", len(result.FilesMissingTests)))
	}
	if v.config.MinCoverage > 0 && result.CoveragePercent < v.config.MinCoverage {
		result.Errors = append(result.Errors, fmt.Sprintf("coverage %.1f%% is below minimum %.1f%%", result.CoveragePercent, v.config.MinCoverage))
	}

	sort.Strings(result.FilesMissingTests)
	sort.Strings(result.Errors)

	return result, nil
}

func discoverTestsForSource(sf *models.SourceFile) []discoveredTest {
	if sf == nil {
		return nil
	}

	language := scanner.NormalizeLanguage(sf.Language)
	candidates := make([]discoveredTest, 0)

	if language == "rust" && hasInlineRustTests(sf.Path) {
		candidates = append(candidates, discoveredTest{path: sf.Path, inline: true})
	}

	for _, candidate := range candidateTestPaths(sf, language) {
		if _, err := os.Stat(candidate); err == nil {
			candidates = append(candidates, discoveredTest{path: candidate})
		}
	}

	return dedupeDiscoveredTests(candidates)
}

func candidateTestPaths(sf *models.SourceFile, language string) []string {
	adapter := adapters.DefaultRegistry().GetAdapter(sf.Language)
	paths := make([]string, 0, 8)
	if adapter != nil {
		paths = append(paths, adapter.GenerateTestPath(sf.Path, ""))
	}

	dir := filepath.Dir(sf.Path)
	base := filepath.Base(sf.Path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	switch language {
	case "go":
		paths = append(paths, filepath.Join(dir, name+"_test.go"))
	case "python":
		paths = append(paths,
			filepath.Join(dir, "test_"+name+".py"),
			filepath.Join(dir, name+"_test.py"),
			filepath.Join(filepath.Dir(dir), "tests", "test_"+name+".py"),
		)
	case "javascript", "typescript":
		paths = append(paths,
			filepath.Join(dir, name+".test"+ext),
			filepath.Join(dir, name+".spec"+ext),
			filepath.Join(dir, "__tests__", name+".test"+ext),
			filepath.Join(dir, "__tests__", name+".spec"+ext),
		)
	case "java":
		testName := name + "Test.java"
		paths = append(paths, filepath.Join(dir, testName))
		mainJava := filepath.Join("src", "main", "java")
		if strings.Contains(dir, mainJava) {
			paths = append(paths, filepath.Join(strings.Replace(dir, mainJava, filepath.Join("src", "test", "java"), 1), testName))
		}
	case "rust":
		paths = append(paths,
			filepath.Join(filepath.Dir(dir), "tests", name+"_test.rs"),
			filepath.Join(dir, name+"_test.rs"),
			sf.Path+".test",
		)
	}

	return paths
}

func looksLikeTestFile(language string, path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read discovered test file %s: %w", path, err)
	}

	text := string(content)
	switch scanner.NormalizeLanguage(language) {
	case "go":
		return strings.Contains(text, "func Test") || strings.Contains(text, "func Benchmark") || strings.Contains(text, "func Fuzz"), nil
	case "python":
		return strings.Contains(text, "def test_") || strings.Contains(text, "class Test"), nil
	case "javascript", "typescript":
		re := regexp.MustCompile(`\b(describe|it|test)\s*(?:\.|\()`)
		return re.MatchString(text), nil
	case "java":
		return strings.Contains(text, "@Test"), nil
	case "rust":
		return strings.Contains(text, "#[test]"), nil
	default:
		return len(strings.TrimSpace(text)) > 0, nil
	}
}

func hasInlineRustTests(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	text := string(content)
	return strings.Contains(text, "#[cfg(test)]") || strings.Contains(text, "#[test]")
}

func dedupeDiscoveredTests(discovered []discoveredTest) []discoveredTest {
	if len(discovered) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(discovered))
	result := make([]discoveredTest, 0, len(discovered))
	for _, item := range discovered {
		key := item.path
		if item.inline {
			key = item.path + "#inline"
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}

	return result
}

func executionRoot(defaultPath string, sourcePath string, discovered []discoveredTest) string {
	paths := []string{defaultPath, filepath.Dir(sourcePath)}
	for _, item := range discovered {
		if item.inline {
			continue
		}
		paths = append(paths, filepath.Dir(item.path))
	}

	root := commonAncestor(paths...)
	if root == "" {
		return defaultPath
	}
	return root
}

func mergeExecutionRoot(current string, next string) string {
	if current == "" {
		return next
	}
	if next == "" {
		return current
	}
	return commonAncestor(current, next)
}

func commonAncestor(paths ...string) string {
	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}

		absolute, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		filtered = append(filtered, absolute)
	}

	if len(filtered) == 0 {
		return ""
	}

	base := filtered[0]
	for _, candidate := range filtered[1:] {
		base = commonAncestorPair(base, candidate)
		if base == "" {
			return ""
		}
	}

	return base
}

func commonAncestorPair(left string, right string) string {
	leftParts := splitPath(left)
	rightParts := splitPath(right)
	limit := len(leftParts)
	if len(rightParts) < limit {
		limit = len(rightParts)
	}

	shared := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		if leftParts[i] != rightParts[i] {
			break
		}
		shared = append(shared, leftParts[i])
	}

	if len(shared) == 0 {
		return ""
	}
	return filepath.Join(shared...)
}

func splitPath(path string) []string {
	cleaned := filepath.Clean(path)
	volume := filepath.VolumeName(cleaned)
	trimmed := strings.TrimPrefix(cleaned, volume)
	trimmed = strings.TrimPrefix(trimmed, string(filepath.Separator))
	parts := []string{}
	if volume != "" {
		parts = append(parts, volume+string(filepath.Separator))
	} else if strings.HasPrefix(cleaned, string(filepath.Separator)) {
		parts = append(parts, string(filepath.Separator))
	}
	if trimmed == "" {
		return parts
	}
	return append(parts, strings.Split(trimmed, string(filepath.Separator))...)
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
