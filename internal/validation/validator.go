/*
Package validation provides test validation and execution functionality.
*/
package validation

import (
	"os"

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

// NewValidator creates a new validator
func NewValidator(config Config) *Validator {
	return &Validator{
		config: config,
	}
}

// Validate validates tests for the given source files
func (v *Validator) Validate(path string, sourceFiles []*models.SourceFile) (*Result, error) {
	result := &Result{
		FilesMissingTests: make([]string, 0),
		Errors:            make([]string, 0),
	}
	coveredLanguages := make(map[string]bool)

	// For now, a simplified validation that checks for test file existence
	for _, sf := range sourceFiles {
		hasTest := checkTestFileExists(sf)
		if hasTest {
			result.FilesWithTests++
			coveredLanguages[scanner.NormalizeLanguage(sf.Language)] = true
		} else {
			result.FilesMissingTests = append(result.FilesMissingTests, sf.Path)
		}
	}

	// Calculate approximate coverage
	total := len(sourceFiles)
	if total > 0 {
		result.CoveragePercent = float64(result.FilesWithTests) / float64(total) * 100
	}

	parser := NewCoverageParser()
	coverageSamples := 0
	for language := range coveredLanguages {
		adapter := adapters.DefaultRegistry().GetAdapter(language)
		if adapter == nil {
			continue
		}
		testResults, err := adapter.RunTests(path)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			continue
		}
		result.TestsPassed += testResults.PassedCount
		result.TestsFailed += testResults.FailedCount
		if testResults.Output != "" {
			if coverage := parser.ParseCoverage(testResults.Output, language); coverage > 0 {
				result.CoveragePercent = coverage
				coverageSamples++
			}
		}
	}

	if coverageSamples == 0 && total == 0 {
		result.CoveragePercent = 0
	}

	return result, nil
}

// checkTestFileExists checks if a test file exists for the source file
func checkTestFileExists(sf *models.SourceFile) bool {
	adapter := adapters.DefaultRegistry().GetAdapter(sf.Language)
	if adapter == nil {
		return false
	}

	testPath := adapter.GenerateTestPath(sf.Path, "")
	if _, err := os.Stat(testPath); err == nil {
		return true
	}

	return false
}
