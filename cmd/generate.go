package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/princepal9120/testgen-cli/internal/ui"
	"github.com/princepal9120/testgen-cli/pkg/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CLI output styles
var (
	successMark = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓")
	errorMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗")
	warnMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("⚠")
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

var (
	// generate command flags
	genPath           string
	genFile           string
	genTypes          []string
	genFramework      string
	genOutput         string
	genRecursive      bool
	genParallel       int
	genDryRun         bool
	genValidate       bool
	genOutputFormat   string
	genIncludePattern string
	genExcludePattern string
	genBatchSize      int
	genReportUsage    bool
	genInteractive    bool
	genEmitPatch      bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate tests for source files",
	Long: `Generate tests for specified source files or directories.

TestGen analyzes your source code, extracts function definitions, and
generates comprehensive tests using AI. Tests follow language-specific
conventions and best practices.

Test Types:
  unit         - Basic unit tests covering happy path and common errors
  edge-cases   - Boundary conditions, nulls, extremes  
  negative     - Exception paths, invalid inputs
  table-driven - Parameterized tests (Go idiom)
  integration  - Tests with mocked external dependencies

Examples:
  # Generate unit tests for a single file
  testgen generate --file=./src/utils.py --type=unit

  # Generate multiple test types for a directory
  testgen generate --path=./src --type=unit,edge-cases --recursive

  # Preview without writing files
  testgen generate --path=./src --dry-run

  # Generate and validate tests
  testgen generate --path=./src --validate`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Path/file flags
	generateCmd.Flags().StringVarP(&genPath, "path", "p", "", "source directory to generate tests for")
	generateCmd.Flags().StringVar(&genFile, "file", "", "single source file to generate tests for")

	// Test configuration
	generateCmd.Flags().StringSliceVarP(&genTypes, "type", "t", []string{"unit"}, "test types: unit, edge-cases, negative, table-driven, integration")
	generateCmd.Flags().StringVarP(&genFramework, "framework", "f", "", "target test framework (auto-detected by default)")
	generateCmd.Flags().StringVarP(&genOutput, "output", "o", "", "output directory for generated tests")

	// Processing options
	generateCmd.Flags().BoolVarP(&genRecursive, "recursive", "r", false, "process directories recursively")
	generateCmd.Flags().IntVarP(&genParallel, "parallel", "j", 2, "number of parallel workers")
	generateCmd.Flags().IntVar(&genBatchSize, "batch-size", 5, "batch size for API requests")

	// Output options
	generateCmd.Flags().BoolVar(&genDryRun, "dry-run", false, "preview output without writing files")
	generateCmd.Flags().BoolVar(&genValidate, "validate", false, "run generated tests after creation")
	generateCmd.Flags().StringVar(&genOutputFormat, "output-format", "text", "output format: text, json")
	generateCmd.Flags().BoolVar(&genEmitPatch, "emit-patch", false, "include structured patch operations in shared/json output")

	// Filtering options
	generateCmd.Flags().StringVar(&genIncludePattern, "include-pattern", "", "glob pattern for files to include")
	generateCmd.Flags().StringVar(&genExcludePattern, "exclude-pattern", "", "glob pattern for files to exclude")

	// Reporting
	generateCmd.Flags().BoolVar(&genReportUsage, "report-usage", false, "generate usage/cost report")

	// Interactive mode
	generateCmd.Flags().BoolVarP(&genInteractive, "interactive", "i", false, "show interactive results view after generation")

	// Bind to viper
	viper.BindPFlag("generation.parallel_workers", generateCmd.Flags().Lookup("parallel"))
	viper.BindPFlag("generation.batch_size", generateCmd.Flags().Lookup("batch-size"))
}

func runGenerate(cmd *cobra.Command, args []string) error {
	log := GetLogger()

	// Validate inputs
	if genPath == "" && genFile == "" {
		return fmt.Errorf("either --path or --file is required")
	}

	// Check API key early (non-quiet mode shows helpful error)
	provider := viper.GetString("llm.provider")
	if provider == "" {
		provider = "anthropic" // default
	}
	apiKey := getAPIKeyForProvider(provider)
	if apiKey == "" && !quiet && genOutputFormat != "json" {
		ui.ShowAPIKeyError(provider)
		return fmt.Errorf("API key not configured for %s", provider)
	}

	// Determine target path
	targetPath := genPath
	if genFile != "" {
		targetPath = genFile
	}

	// Make path absolute
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	log.Info("starting test generation",
		slog.String("path", absPath),
		slog.Any("types", genTypes),
		slog.Bool("recursive", genRecursive),
		slog.Bool("dry-run", genDryRun),
	)

	service := app.NewService()
	response, err := service.Generate(context.Background(), app.GenerateRequest{
		Path:           genPath,
		File:           genFile,
		Recursive:      genRecursive,
		IncludePattern: genIncludePattern,
		ExcludePattern: genExcludePattern,
		TestTypes:      genTypes,
		Framework:      genFramework,
		OutputDir:      genOutput,
		DryRun:         genDryRun,
		Validate:       genValidate,
		BatchSize:      genBatchSize,
		Parallelism:    genParallel,
		Provider:       viper.GetString("llm.provider"),
		EmitPatch:      genEmitPatch,
	})
	if err != nil {
		return err
	}
	results := response.Results

	log.Info("found source files",
		slog.Int("count", len(response.SourceFiles)),
		slog.String("path", response.TargetPath),
	)

	// Show interactive results or text output
	if genInteractive && !genDryRun && genOutputFormat != "json" {
		log.Info("generation complete", slog.Int("files", len(results)))
		return ui.ShowResults(results)
	}

	// Output results
	if err := outputResults(response, genOutputFormat, genDryRun); err != nil {
		return fmt.Errorf("failed to output results: %w", err)
	}

	// Summary
	successCount := response.SuccessCount
	errorCount := response.ErrorCount

	log.Info("generation complete",
		slog.Int("success", successCount),
		slog.Int("errors", errorCount),
		slog.Int("total", len(results)),
	)

	// Show TUI banner (non-quiet, non-json mode)
	if !quiet && genOutputFormat != "json" {
		if errorCount > 0 {
			ui.ShowError(
				fmt.Sprintf("%d file(s) failed to generate tests", errorCount),
				"Run with --verbose for details",
			)
			return fmt.Errorf("%d file(s) failed to generate tests", errorCount)
		}

		funcsCount := 0
		for _, r := range results {
			funcsCount += len(r.FunctionsTested)
		}
		ui.ShowSuccess(ui.SuccessStats{
			FilesProcessed: len(results),
			TestsGenerated: successCount,
			FunctionsFound: funcsCount,
		})
		return nil
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) failed to generate tests", errorCount)
	}

	return nil
}

func outputResults(response *app.GenerateResponse, format string, dryRun bool) error {
	switch strings.ToLower(format) {
	case "json":
		return outputJSON(response)
	default:
		return outputText(response.Results, dryRun)
	}
}

func outputJSON(response *app.GenerateResponse) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

func outputText(results []*models.GenerationResult, dryRun bool) error {
	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("%s %s: %v\n", errorMark, r.SourceFile.Path, r.Error)
			continue
		}

		if dryRun && r.TestCode != "" {
			fmt.Printf("\n--- %s (generated test) ---\n", r.SourceFile.Path)
			fmt.Println(r.TestCode)
			fmt.Println()
		} else if r.TestPath != "" {
			funcInfo := dimStyle.Render(fmt.Sprintf("(%d functions)", len(r.FunctionsTested)))
			fmt.Printf("%s %s → %s %s\n", successMark, r.SourceFile.Path, r.TestPath, funcInfo)
		}
	}
	return nil
}

func getAPIKeyForProvider(provider string) string {
	switch strings.ToLower(provider) {
	case "openai":
		return os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "gemini":
		key := os.Getenv("GEMINI_API_KEY")
		if key == "" {
			key = os.Getenv("GOOGLE_API_KEY")
		}
		return key
	case "groq":
		return os.Getenv("GROQ_API_KEY")
	default:
		return ""
	}
}
