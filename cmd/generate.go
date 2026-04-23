package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/princepal9120/testgen-cli/internal/llm"
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
	genRequestFile    string
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
	generateCmd.Flags().StringVar(&genRequestFile, "request-file", "", "read a machine request from JSON file ('-' reads stdin)")

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
	machineMode := isGenerateMachineMode()
	outputFormat := effectiveGenerateOutputFormat()
	if cmd != nil && machineMode {
		previousQuiet := quiet
		quiet = true
		defer func() { quiet = previousQuiet }()
		initLogger()
		root := cmd.Root()
		if root != nil {
			root.SilenceErrors = true
			root.SilenceUsage = true
		}
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}
	log := GetLogger()

	req, err := buildGenerateRequest(cmd)
	if err != nil {
		if outputFormat == "json" {
			_ = outputJSON(app.NewGenerateFailureResponse(req, err, appTargetPathHint(req)))
		}
		return err
	}

	if req.Provider == "" {
		req.Provider = viper.GetString("llm.provider")
	}
	if req.Provider == "" {
		req.Provider = "anthropic"
	}

	// Check API key early only for human/text mode. JSON/machine mode should
	// let the service produce a structured envelope so dry-run/no-definition
	// flows can still succeed without requiring credentials up front.
	apiKey := getAPIKeyForProvider(req.Provider)
	if apiKey == "" && outputFormat != "json" {
		err = fmt.Errorf("%w for %s", llm.ErrNoAPIKey, req.Provider)
		if !quiet {
			ui.ShowAPIKeyError(req.Provider)
		}
		return err
	}

	log.Info("starting test generation",
		slog.String("path", appTargetPathHint(req)),
		slog.Any("types", req.TestTypes),
		slog.Bool("recursive", req.Recursive),
		slog.Bool("dry-run", req.ResolvedDryRun()),
		slog.String("request_id", req.RequestID),
	)

	service := app.NewService()
	response, err := service.Generate(context.Background(), req)
	if err != nil {
		if outputFormat == "json" {
			_ = outputJSON(app.NewGenerateFailureResponse(req, err, appTargetPathHint(req)))
		}
		return err
	}
	results := response.Results

	log.Info("found source files",
		slog.Int("count", len(response.SourceFiles)),
		slog.String("path", response.TargetPath),
	)

	// Show interactive results or text output
	if genInteractive && !response.DryRun && outputFormat != "json" {
		log.Info("generation complete", slog.Int("files", len(results)))
		return ui.ShowResults(results)
	}

	// Output results
	if err := outputResults(response, outputFormat, response.DryRun); err != nil {
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
	if !quiet && outputFormat != "json" {
		if errorCount > 0 {
			ui.ShowError(
				fmt.Sprintf("%d file(s) failed to generate tests", errorCount),
				"Run with --verbose for details",
			)
			return machineError(response, fmt.Sprintf("%d file(s) failed to generate tests", errorCount))
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
		return machineError(response, fmt.Sprintf("%d file(s) failed to generate tests", errorCount))
	}

	return nil
}

func isGenerateMachineMode() bool {
	return strings.EqualFold(genOutputFormat, "json") || genRequestFile != ""
}

func effectiveGenerateOutputFormat() string {
	if isGenerateMachineMode() {
		return "json"
	}
	return strings.ToLower(genOutputFormat)
}

func buildGenerateRequest(cmd *cobra.Command) (app.GenerateRequest, error) {
	req := app.GenerateRequest{}
	if genRequestFile != "" {
		loaded, err := loadGenerateRequest(genRequestFile)
		if err != nil {
			return req, err
		}
		req = loaded
	}

	shouldOverlay := func(flagName string) bool {
		return genRequestFile == "" || (cmd != nil && cmd.Flags().Changed(flagName))
	}

	if shouldOverlay("path") {
		req.Path = genPath
	}
	if shouldOverlay("file") {
		req.File = genFile
	}
	if shouldOverlay("recursive") {
		req.Recursive = genRecursive
	}
	if shouldOverlay("include-pattern") {
		req.IncludePattern = genIncludePattern
	}
	if shouldOverlay("exclude-pattern") {
		req.ExcludePattern = genExcludePattern
	}
	if shouldOverlay("type") {
		req.TestTypes = append([]string(nil), genTypes...)
	}
	if shouldOverlay("framework") {
		req.Framework = genFramework
	}
	if shouldOverlay("output") {
		req.OutputDir = genOutput
	}
	if shouldOverlay("dry-run") {
		req.DryRun = genDryRun
	}
	if shouldOverlay("validate") {
		req.Validate = genValidate
	}
	if shouldOverlay("batch-size") {
		req.BatchSize = genBatchSize
	}
	if shouldOverlay("report-usage") {
		req.ReportUsage = genReportUsage
	}
	if shouldOverlay("parallel") {
		req.Parallelism = genParallel
	}
	if shouldOverlay("report-usage") {
		req.ReportUsage = genReportUsage
	}
	if shouldOverlay("emit-patch") {
		req.EmitPatch = genEmitPatch
	}

	return req, nil
}

func loadGenerateRequest(path string) (app.GenerateRequest, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return app.GenerateRequest{}, fmt.Errorf("failed to read request file: %w", err)
	}

	var payload struct {
		app.GenerateRequest
		Types []string `json:"types"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&payload); err != nil {
		return app.GenerateRequest{}, fmt.Errorf("failed to decode request file: %w", err)
	}
	req := payload.GenerateRequest
	if len(req.TestTypes) == 0 && len(payload.Types) > 0 {
		req.TestTypes = append([]string(nil), payload.Types...)
	}
	return req, nil
}

func machineError(response *app.GenerateResponse, fallback string) error {
	if response != nil && response.Error != "" {
		return fmt.Errorf("%s", response.Error)
	}
	return fmt.Errorf("%s", fallback)
}

func outputResults(response *app.GenerateResponse, format string, dryRun bool) error {
	switch strings.ToLower(format) {
	case "json":
		return outputJSON(response)
	default:
		return outputText(response, dryRun)
	}
}

func outputJSON(response *app.GenerateResponse) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

func appTargetPathHint(req app.GenerateRequest) string {
	if req.File != "" {
		if absPath, err := filepath.Abs(req.File); err == nil {
			return absPath
		}
		return req.File
	}
	if req.Path != "" {
		if absPath, err := filepath.Abs(req.Path); err == nil {
			return absPath
		}
		return req.Path
	}
	return ""
}

func outputText(response *app.GenerateResponse, dryRun bool) error {
	if response == nil {
		return nil
	}

	for _, r := range response.Results {
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
	if response.Usage != nil {
		fmt.Printf("\n--- Usage Report ---\n")
		if response.Usage.Provider != "" {
			fmt.Printf("Provider:        %s\n", response.Usage.Provider)
		}
		if response.Usage.Model != "" {
			fmt.Printf("Model:           %s\n", response.Usage.Model)
		}
		fmt.Printf("Requests:        %d\n", response.Usage.TotalRequests)
		fmt.Printf("Input tokens:    %d\n", response.Usage.InputTokens)
		fmt.Printf("Output tokens:   %d\n", response.Usage.OutputTokens)
		fmt.Printf("Cached tokens:   %d\n", response.Usage.CachedTokens)
		fmt.Printf("Cache hits:      %d\n", response.Usage.CacheHits)
		fmt.Printf("Cache misses:    %d\n", response.Usage.CacheMisses)
		fmt.Printf("Cache hit rate:  %.2f%%\n", response.Usage.CacheHitRate*100)
		fmt.Printf("Estimated cost:  $%.4f USD\n", response.Usage.EstimatedCostUSD)
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
