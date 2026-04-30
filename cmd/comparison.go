package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type comparisonResponse struct {
	APIVersion string                 `json:"api_version"`
	Success    bool                   `json:"success"`
	Path       string                 `json:"path,omitempty"`
	Cost       *app.AnalyzeResponse   `json:"cost,omitempty"`
	Summary    string                 `json:"summary"`
	Rows       []comparisonRow        `json:"rows"`
	Commands   []comparisonCommand    `json:"commands"`
	NextSteps  []string               `json:"next_steps"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type comparisonRow struct {
	Dimension string `json:"dimension"`
	PlainLLM  string `json:"plain_llm"`
	TestGen   string `json:"testgen_skill"`
}

type comparisonCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

var (
	comparisonPath         string
	comparisonRecursive    bool
	comparisonProvider     string
	comparisonModel        string
	comparisonBatchSize    int
	comparisonOutputFormat string
)

var comparisonCmd = &cobra.Command{
	Use:     "comparison",
	Aliases: []string{"compare", "vs"},
	Short:   "Compare plain LLM test generation with the TestGen skill workflow",
	Long: `Compare a generic LLM prompt flow against TestGen's agent skill workflow.

The report is designed for product demos, READMEs, and agent evaluation. When a
path is supplied, it also includes an offline TestGen cost estimate for that
codebase.`,
	Example: `  testgen comparison --path=./src
  testgen comparison --path=./src --provider=gemini --output-format=json`,
	RunE: runComparison,
}

func init() {
	rootCmd.AddCommand(comparisonCmd)

	comparisonCmd.Flags().StringVarP(&comparisonPath, "path", "p", ".", "file or directory to include in the comparison")
	comparisonCmd.Flags().BoolVarP(&comparisonRecursive, "recursive", "r", true, "analyze directories recursively")
	comparisonCmd.Flags().StringVar(&comparisonProvider, "provider", "", "LLM provider for pricing: anthropic, openai, gemini, groq")
	comparisonCmd.Flags().StringVar(&comparisonModel, "model", "", "model name for pricing; provider default is used when omitted")
	comparisonCmd.Flags().IntVar(&comparisonBatchSize, "batch-size", 5, "estimated batch size for generation requests")
	comparisonCmd.Flags().StringVar(&comparisonOutputFormat, "output-format", "text", "output format: text, json")
}

func runComparison(cmd *cobra.Command, args []string) error {
	machineMode := strings.EqualFold(comparisonOutputFormat, "json")
	if machineMode {
		previousQuiet := quiet
		quiet = true
		defer func() { quiet = previousQuiet }()
		initLogger()
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}

	provider := comparisonProvider
	if provider == "" {
		provider = viper.GetString("llm.provider")
	}
	if provider == "" {
		provider = "anthropic"
	}

	costReq := app.AnalyzeRequest{
		Path:         comparisonPath,
		Recursive:    comparisonRecursive,
		CostEstimate: true,
		Provider:     provider,
		Model:        comparisonModel,
		BatchSize:    comparisonBatchSize,
		Detail:       "summary",
	}
	cost, err := app.NewService().Analyze(context.Background(), costReq)
	if err != nil {
		if machineMode {
			resp := comparisonResponse{
				APIVersion: app.APIVersion,
				Success:    false,
				Path:       comparisonPath,
				Summary:    err.Error(),
				Rows:       defaultComparisonRows(),
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			_ = encoder.Encode(resp)
		}
		return err
	}

	resp := comparisonResponse{
		APIVersion: app.APIVersion,
		Success:    true,
		Path:       cost.Path,
		Cost:       cost,
		Summary:    "Plain LLM prompts can generate tests, but TestGen adds repeatable scanning, cost planning, dry-run patches, usage reporting, validation, and agent-safe write controls.",
		Rows:       defaultComparisonRows(),
		Commands: []comparisonCommand{
			{Name: "Cost", Description: "Estimate before spending tokens", Command: fmt.Sprintf("testgen cost --path=%s --provider=%s", comparisonPath, provider)},
			{Name: "Generate", Description: "Create review-first test patches", Command: fmt.Sprintf("testgen generate --path=%s --recursive --type=unit --dry-run --emit-patch --report-usage", comparisonPath)},
			{Name: "Comparison", Description: "Explain LLM vs TestGen skill", Command: fmt.Sprintf("testgen comparison --path=%s", comparisonPath)},
		},
		NextSteps: []string{
			"Run cost first for the target folder.",
			"Generate dry-run patches and inspect artifacts.",
			"Write and validate tests only after review or explicit approval.",
		},
	}

	if machineMode {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(resp)
	}
	return outputComparison(resp)
}

func defaultComparisonRows() []comparisonRow {
	return []comparisonRow{
		{Dimension: "Repo awareness", PlainLLM: "Depends on pasted context", TestGen: "Scans files, languages, functions, and nearby test style"},
		{Dimension: "Cost control", PlainLLM: "Usually unknown until after the request", TestGen: "Offline estimates through testgen cost before API calls"},
		{Dimension: "Safety", PlainLLM: "May write or overwrite without a structured review step", TestGen: "Dry-run first, patch artifacts, explicit write controls"},
		{Dimension: "Repeatability", PlainLLM: "Prompt quality varies by agent and session", TestGen: "Same engine across Codex, Claude Code, OpenCode, CLI, and MCP"},
		{Dimension: "Validation", PlainLLM: "Manual follow-up", TestGen: "Built-in validate flow and machine-readable results"},
		{Dimension: "Agent integration", PlainLLM: "One-off prompt", TestGen: "Repo-local skill/command plus JSON envelopes for agents"},
	}
}

func outputComparison(resp comparisonResponse) error {
	fmt.Printf("\n=== LLM vs TestGen Skill Comparison ===\n\n")
	fmt.Println(resp.Summary)
	fmt.Printf("\nPath:              %s\n", resp.Path)
	if resp.Cost != nil {
		fmt.Printf("Source files:      %d\n", resp.Cost.TotalFiles)
		fmt.Printf("Est. functions:    %d\n", resp.Cost.TotalFunctions)
		fmt.Printf("Provider/model:    %s / %s\n", resp.Cost.Provider, resp.Cost.Model)
		fmt.Printf("Est. requests:     %d\n", resp.Cost.EstimatedRequests)
		fmt.Printf("Est. tokens:       %d\n", resp.Cost.EstimatedTokens)
		fmt.Printf("Est. cost:         $%.4f USD\n", resp.Cost.EstimatedCost)
	}

	fmt.Printf("\nComparison:\n")
	for _, row := range resp.Rows {
		fmt.Printf("\n- %s\n", row.Dimension)
		fmt.Printf("  Plain LLM: %s\n", row.PlainLLM)
		fmt.Printf("  TestGen:   %s\n", row.TestGen)
	}

	fmt.Printf("\nRecommended commands:\n")
	for _, command := range resp.Commands {
		fmt.Printf("  # %s: %s\n", command.Name, command.Description)
		fmt.Printf("  %s\n", command.Command)
	}

	fmt.Printf("\nNext steps:\n")
	for _, step := range resp.NextSteps {
		fmt.Printf("  - %s\n", step)
	}
	fmt.Println()
	return nil
}
