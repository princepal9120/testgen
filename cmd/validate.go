package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/princepal9120/testgen-cli/internal/validation"
	"github.com/spf13/cobra"
)

var (
	// validate command flags
	valPath          string
	valRecursive     bool
	valMinCoverage   float64
	valFailOnMissing bool
	valReportGaps    bool
	valOutputFormat  string
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate existing tests and coverage",
	Long: `Validate test files and analyze coverage for a codebase.

This command checks that:
  • Test files compile/parse correctly
  • Tests execute successfully
  • Code coverage meets minimum thresholds
  • Source files have corresponding tests

Examples:
  # Check basic validation
  testgen validate --path=./src

  # Enforce minimum coverage
  testgen validate --path=./src --min-coverage=80

  # Fail if any source files lack tests
  testgen validate --path=./src --fail-on-missing-tests

  # Show detailed coverage gaps
  testgen validate --path=./src --report-gaps`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&valPath, "path", "p", ".", "directory to validate")
	validateCmd.Flags().BoolVarP(&valRecursive, "recursive", "r", true, "check recursively")
	validateCmd.Flags().Float64Var(&valMinCoverage, "min-coverage", 0, "minimum coverage percentage (0-100)")
	validateCmd.Flags().BoolVar(&valFailOnMissing, "fail-on-missing-tests", false, "exit with error if tests missing")
	validateCmd.Flags().BoolVar(&valReportGaps, "report-gaps", false, "show coverage gaps per file")
	validateCmd.Flags().StringVar(&valOutputFormat, "output-format", "text", "output format: text, json")
}

func runValidate(cmd *cobra.Command, args []string) error {
	log := GetLogger()

	log.Info("validating tests",
		slog.String("path", valPath),
		slog.Float64("min-coverage", valMinCoverage),
		slog.Bool("recursive", valRecursive),
	)

	service := app.NewService()
	response, err := service.Validate(context.Background(), app.ValidateRequest{
		Path:          valPath,
		Recursive:     valRecursive,
		MinCoverage:   valMinCoverage,
		FailOnMissing: valFailOnMissing,
		ReportGaps:    valReportGaps,
	})
	if err != nil {
		return err
	}
	result := response.Result

	// Output results
	if err := outputValidationResults(result, valOutputFormat); err != nil {
		return err
	}

	// Check thresholds
	if valMinCoverage > 0 && result.CoveragePercent < valMinCoverage {
		return fmt.Errorf("coverage %.1f%% is below minimum %.1f%%", result.CoveragePercent, valMinCoverage)
	}

	if valFailOnMissing && len(result.FilesMissingTests) > 0 {
		return fmt.Errorf("%d file(s) are missing tests", len(result.FilesMissingTests))
	}

	log.Info("validation complete",
		slog.Float64("coverage", result.CoveragePercent),
		slog.Int("files-with-tests", result.FilesWithTests),
		slog.Int("files-missing-tests", len(result.FilesMissingTests)),
	)

	return nil
}

func outputValidationResults(result *validation.Result, format string) error {
	switch strings.ToLower(format) {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	default:
		fmt.Printf("\n=== Validation Results ===\n\n")
		fmt.Printf("Coverage:           %.1f%%\n", result.CoveragePercent)
		fmt.Printf("Files with tests:   %d\n", result.FilesWithTests)
		fmt.Printf("Files missing tests: %d\n", len(result.FilesMissingTests))
		fmt.Printf("Tests passed:       %d\n", result.TestsPassed)
		fmt.Printf("Tests failed:       %d\n", result.TestsFailed)

		if len(result.FilesMissingTests) > 0 && valReportGaps {
			fmt.Printf("\n--- Files Missing Tests ---\n")
			for _, f := range result.FilesMissingTests {
				fmt.Printf("  • %s\n", f)
			}
		}

		if len(result.Errors) > 0 {
			fmt.Printf("\n--- Errors ---\n")
			for _, e := range result.Errors {
				fmt.Printf("  ✗ %s\n", e)
			}
		}
		fmt.Println()
		return nil
	}
}
