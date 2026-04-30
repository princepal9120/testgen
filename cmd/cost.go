package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	costPath         string
	costRecursive    bool
	costDetail       string
	costProvider     string
	costModel        string
	costBatchSize    int
	costOutputFormat string
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Estimate TestGen generation cost before API calls",
	Long: `Estimate test-generation cost for a file or directory without calling an LLM.

This is a convenience command for agents and users who want a direct cost command
instead of remembering the analyze flags. It runs the same offline scanner and
provider-aware estimator as:

  testgen analyze --cost-estimate --output-format json`,
	Example: `  testgen cost --path=./src
  testgen cost --path=./src --provider=gemini --model=gemini-1.5-flash
  testgen cost --path=./src --detail=per-file --output-format=json`,
	RunE: runCost,
}

func init() {
	rootCmd.AddCommand(costCmd)

	costCmd.Flags().StringVarP(&costPath, "path", "p", ".", "file or directory to estimate")
	costCmd.Flags().BoolVarP(&costRecursive, "recursive", "r", true, "analyze directories recursively")
	costCmd.Flags().StringVar(&costDetail, "detail", "summary", "detail level: summary, per-file")
	costCmd.Flags().StringVar(&costProvider, "provider", "", "LLM provider for pricing: anthropic, openai, gemini, groq")
	costCmd.Flags().StringVar(&costModel, "model", "", "model name for pricing; provider default is used when omitted")
	costCmd.Flags().IntVar(&costBatchSize, "batch-size", 5, "estimated batch size for generation requests")
	costCmd.Flags().StringVar(&costOutputFormat, "output-format", "text", "output format: text, json")
}

func runCost(cmd *cobra.Command, args []string) error {
	machineMode := strings.EqualFold(costOutputFormat, "json")
	if machineMode {
		previousQuiet := quiet
		quiet = true
		defer func() { quiet = previousQuiet }()
		initLogger()
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}

	provider := costProvider
	if provider == "" {
		provider = viper.GetString("llm.provider")
	}
	if provider == "" {
		provider = "anthropic"
	}

	req := app.AnalyzeRequest{
		Path:         costPath,
		Recursive:    costRecursive,
		CostEstimate: true,
		Provider:     provider,
		Model:        costModel,
		BatchSize:    costBatchSize,
		Detail:       costDetail,
	}

	result, err := app.NewService().Analyze(context.Background(), req)
	if err != nil {
		if machineMode {
			resp := app.NewAnalyzeFailureResponse(req, err, costPath)
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			_ = encoder.Encode(resp)
		}
		return err
	}

	if machineMode {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	return outputCostSummary(result, costDetail)
}

func outputCostSummary(result *app.AnalyzeResponse, detail string) error {
	fmt.Printf("\n=== TestGen Cost Estimate ===\n\n")
	fmt.Printf("Path:              %s\n", result.Path)
	fmt.Printf("Provider:          %s\n", result.Provider)
	fmt.Printf("Model:             %s\n", result.Model)
	fmt.Printf("Source files:      %d\n", result.TotalFiles)
	fmt.Printf("Est. functions:    %d\n", result.TotalFunctions)
	fmt.Printf("Est. requests:     %d\n", result.EstimatedRequests)
	fmt.Printf("Est. input tokens: %d\n", result.EstimatedInputTokens)
	fmt.Printf("Est. output tokens:%d\n", result.EstimatedOutputTokens)
	fmt.Printf("Est. total tokens: %d\n", result.EstimatedTokens)
	fmt.Printf("Est. cost:         $%.4f USD\n", result.EstimatedCost)

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}

	if len(result.ByLanguage) > 0 {
		fmt.Printf("\nBy language:\n")
		languages := make([]string, 0, len(result.ByLanguage))
		for lang := range result.ByLanguage {
			languages = append(languages, lang)
		}
		sort.Strings(languages)
		for _, lang := range languages {
			stats := result.ByLanguage[lang]
			fmt.Printf("  - %s: %d files, %d lines, ~%d functions\n", lang, stats.Files, stats.Lines, stats.Functions)
		}
	}

	if detail == "per-file" && len(result.Files) > 0 {
		fmt.Printf("\nPer-file estimate:\n")
		for _, f := range result.Files {
			fmt.Printf("  - %s: ~%d tokens, $%.4f\n", f.Path, f.Tokens, f.EstimatedCost)
		}
	}

	fmt.Println()
	return nil
}
