package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/spf13/cobra"
)

type capabilityCommand struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Flags       []string `json:"flags,omitempty"`
	WritesFiles bool     `json:"writes_files"`
}

type capabilitiesResponse struct {
	APIVersion        string              `json:"api_version"`
	Success           bool                `json:"success"`
	SchemaVersion     string              `json:"schema_version"`
	Languages         []languageInfo      `json:"languages"`
	Commands          []capabilityCommand `json:"commands"`
	DryRunSupported   bool                `json:"dry_run_supported"`
	PatchSupported    bool                `json:"patch_supported"`
	ValidationSupport map[string]string   `json:"validation_support"`
	Providers         []string            `json:"providers"`
	Limitations       []string            `json:"limitations"`
}

var capabilitiesOutputFormat string

var capabilitiesCmd = &cobra.Command{
	Use:     "capabilities",
	Aliases: []string{"caps"},
	Short:   "Print an agent-readable TestGen capability manifest",
	Long: `Print a stable manifest describing TestGen commands, supported languages,
write behavior, validation support, providers, and known limitations.`,
	Example: `  testgen capabilities
  testgen capabilities --output-format=json`,
	RunE: runCapabilities,
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
	capabilitiesCmd.Flags().StringVar(&capabilitiesOutputFormat, "output-format", "text", "output format: text, json")
}

func runCapabilities(cmd *cobra.Command, args []string) error {
	machineMode := strings.EqualFold(capabilitiesOutputFormat, "json")
	if machineMode {
		previousQuiet := quiet
		quiet = true
		defer func() { quiet = previousQuiet }()
		initLogger()
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}

	resp := buildCapabilitiesResponse()
	if machineMode {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(resp)
	}
	return outputCapabilities(resp)
}

func buildCapabilitiesResponse() capabilitiesResponse {
	languages := supportedLanguageInfo()
	validation := make(map[string]string, len(languages))
	for _, lang := range languages {
		validation[lang.Language] = lang.ValidationCommand
	}

	return capabilitiesResponse{
		APIVersion:        app.APIVersion,
		Success:           true,
		SchemaVersion:     "capabilities.v1",
		Languages:         languages,
		Commands:          capabilityCommands(),
		DryRunSupported:   true,
		PatchSupported:    true,
		ValidationSupport: validation,
		Providers:         []string{"anthropic", "openai", "gemini", "groq"},
		Limitations: []string{
			"Framework detection is best-effort and repo-marker based.",
			"Generation requires a configured provider API key unless dry-run exits before model calls.",
			"Generated tests should be reviewed before writing or committing.",
		},
	}
}

func capabilityCommands() []capabilityCommand {
	return []capabilityCommand{
		{Name: "analyze", Description: "Analyze source files and estimate generation cost.", Flags: []string{"--path", "--recursive", "--cost-estimate", "--detail", "--output-format"}, WritesFiles: false},
		{Name: "languages", Description: "List supported languages and frameworks.", Flags: []string{"--output-format"}, WritesFiles: false},
		{Name: "capabilities", Description: "Print this capability manifest.", Flags: []string{"--output-format"}, WritesFiles: false},
		{Name: "doctor", Description: "Report repo readiness for TestGen.", Flags: []string{"--path", "--recursive", "--output-format"}, WritesFiles: false},
		{Name: "generate", Description: "Generate tests, optionally as a dry-run patch.", Flags: []string{"--file", "--path", "--recursive", "--type", "--framework", "--output", "--dry-run", "--emit-patch", "--validate", "--output-format", "--request-file"}, WritesFiles: true},
		{Name: "validate", Description: "Validate existing tests and coverage metadata.", Flags: []string{"--path", "--recursive", "--min-coverage", "--fail-on-missing-tests", "--report-gaps", "--output-format"}, WritesFiles: false},
		{Name: "cost", Description: "Friendly alias for cost-focused analyze output.", Flags: []string{"--path", "--recursive", "--detail", "--output-format"}, WritesFiles: false},
		{Name: "comparison", Description: "Compare plain LLM generation with TestGen-guided generation.", Flags: []string{"--path", "--provider", "--output-format"}, WritesFiles: false},
		{Name: "mcp", Description: "Run the stdio MCP server.", WritesFiles: false},
	}
}

func outputCapabilities(resp capabilitiesResponse) error {
	fmt.Printf("\n=== TestGen Capabilities ===\n\n")
	fmt.Printf("Schema: %s\n", resp.SchemaVersion)
	fmt.Printf("Languages: %d\n", len(resp.Languages))
	fmt.Printf("Commands: %d\n", len(resp.Commands))
	fmt.Printf("Dry-run patches: %t\n", resp.DryRunSupported && resp.PatchSupported)
	fmt.Printf("Providers: %s\n\n", strings.Join(resp.Providers, ", "))
	fmt.Println("Suggested agent flow:")
	fmt.Println("  testgen doctor --path=. --output-format=json")
	fmt.Println("  testgen analyze --path=./src --cost-estimate --output-format=json")
	fmt.Println("  testgen generate --path=./src --recursive --dry-run --emit-patch --output-format=json")
	fmt.Println()
	return nil
}
