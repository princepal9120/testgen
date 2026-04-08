package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/spf13/cobra"
)

var (
	// analyze command flags
	anaPath         string
	anaCostEstimate bool
	anaDetail       string
	anaRecursive    bool
	anaOutputFormat string
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze codebase for test generation cost estimation",
	Long: `Analyze source files to estimate test generation costs and complexity.

This command scans your codebase and provides:
  • Estimated token usage for LLM API calls
  • Approximate cost in USD
  • File and function counts per language
  • Complexity metrics

Examples:
  # Get cost estimate for a directory
  testgen analyze --path=./src --cost-estimate

  # Detailed per-file analysis
  testgen analyze --path=./src --cost-estimate --detail=per-file

  # Summary only
  testgen analyze --path=./src --detail=summary`,
	RunE: runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&anaPath, "path", "p", ".", "directory to analyze")
	analyzeCmd.Flags().BoolVar(&anaCostEstimate, "cost-estimate", false, "show estimated API costs")
	analyzeCmd.Flags().StringVar(&anaDetail, "detail", "summary", "detail level: summary, per-file, per-function")
	analyzeCmd.Flags().BoolVarP(&anaRecursive, "recursive", "r", true, "analyze recursively")
	analyzeCmd.Flags().StringVar(&anaOutputFormat, "output-format", "text", "output format: text, json")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	log := GetLogger()

	log.Info("analyzing codebase",
		slog.String("path", anaPath),
		slog.Bool("cost-estimate", anaCostEstimate),
		slog.String("detail", anaDetail),
	)

	service := app.NewService()
	result, err := service.Analyze(context.Background(), app.AnalyzeRequest{
		Path:         anaPath,
		Recursive:    anaRecursive,
		CostEstimate: anaCostEstimate,
		Detail:       anaDetail,
	})
	if err != nil {
		return err
	}

	// Output results
	return outputAnalysisResults(result, anaOutputFormat, anaDetail)
}

func outputAnalysisResults(result *app.AnalyzeResponse, format, detail string) error {
	switch strings.ToLower(format) {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	default:
		fmt.Printf("\n=== Codebase Analysis ===\n\n")
		fmt.Printf("Path:            %s\n", result.Path)
		fmt.Printf("Total files:     %d\n", result.TotalFiles)
		fmt.Printf("Total lines:     %d\n", result.TotalLines)
		fmt.Printf("Est. functions:  %d\n", result.TotalFunctions)

		if len(result.ByLanguage) > 0 {
			fmt.Printf("\n--- By Language ---\n")
			for lang, stats := range result.ByLanguage {
				fmt.Printf("  %s: %d files, %d lines, ~%d functions\n",
					lang, stats.Files, stats.Lines, stats.Functions)
			}
		}

		if result.EstimatedTokens > 0 {
			fmt.Printf("\n--- Cost Estimate ---\n")
			fmt.Printf("Estimated tokens: %d\n", result.EstimatedTokens)
			fmt.Printf("Estimated cost:   $%.2f USD\n", result.EstimatedCost)
		}

		if detail == "per-file" && len(result.Files) > 0 {
			fmt.Printf("\n--- Per-File Details ---\n")
			for _, f := range result.Files {
				fmt.Printf("  %s (%s): %d lines, ~%d functions\n",
					f.Path, f.Language, f.Lines, f.Functions)
			}
		}

		fmt.Println()
		return nil
	}
}
