package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/princepal9120/testgen-cli/internal/scanner"
	"github.com/spf13/cobra"
)

type languageInfo struct {
	Language          string   `json:"language"`
	Aliases           []string `json:"aliases,omitempty"`
	Extensions        []string `json:"extensions"`
	DefaultFramework  string   `json:"default_framework"`
	Frameworks        []string `json:"frameworks"`
	Adapter           string   `json:"adapter"`
	Parser            string   `json:"parser"`
	ValidationCommand string   `json:"validation_command,omitempty"`
}

type languagesResponse struct {
	APIVersion string         `json:"api_version"`
	Success    bool           `json:"success"`
	Count      int            `json:"count"`
	Languages  []languageInfo `json:"languages"`
}

var languagesOutputFormat string

var languagesCmd = &cobra.Command{
	Use:     "languages",
	Aliases: []string{"langs", "frameworks"},
	Short:   "List supported languages, extensions, and test frameworks",
	Long: `List every TestGen-supported language family with file extensions,
default framework, supported frameworks, and adapter maturity.

Use JSON output when an agent needs a machine-readable capability manifest.`,
	Example: `  testgen languages
  testgen languages --output-format=json`,
	RunE: runLanguages,
}

func init() {
	rootCmd.AddCommand(languagesCmd)
	languagesCmd.Flags().StringVar(&languagesOutputFormat, "output-format", "text", "output format: text, json")
}

func runLanguages(cmd *cobra.Command, args []string) error {
	machineMode := strings.EqualFold(languagesOutputFormat, "json")
	if machineMode {
		previousQuiet := quiet
		quiet = true
		defer func() { quiet = previousQuiet }()
		initLogger()
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}

	resp := languagesResponse{
		APIVersion: app.APIVersion,
		Success:    true,
		Languages:  supportedLanguageInfo(),
	}
	resp.Count = len(resp.Languages)

	if machineMode {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(resp)
	}
	return outputLanguages(resp)
}

func supportedLanguageInfo() []languageInfo {
	registry := adapters.DefaultRegistry()
	languages := registry.ListLanguages()
	sort.Strings(languages)

	infos := make([]languageInfo, 0, len(languages)+1)
	for _, lang := range languages {
		adapter := registry.GetAdapter(lang)
		if adapter == nil {
			continue
		}
		info := languageInfo{
			Language:         lang,
			Aliases:          languageAliases(lang),
			Extensions:       scanner.GetExtensionsForLanguage(lang),
			DefaultFramework: adapter.GetDefaultFramework(),
			Frameworks:       adapter.GetSupportedFrameworks(),
			Adapter:          adapterKind(lang),
			Parser:           parserKind(lang),
		}
		info.ValidationCommand = validationCommand(lang)
		infos = append(infos, info)
	}

	// Surface TypeScript as first-class support even though it intentionally uses
	// the JavaScript adapter internally.
	if js := registry.GetAdapter(scanner.LangJavaScript); js != nil {
		infos = append(infos, languageInfo{
			Language:          scanner.LangTypeScript,
			Aliases:           languageAliases(scanner.LangTypeScript),
			Extensions:        scanner.GetExtensionsForLanguage(scanner.LangTypeScript),
			DefaultFramework:  js.GetDefaultFramework(),
			Frameworks:        js.GetSupportedFrameworks(),
			Adapter:           "javascript",
			Parser:            "javascript-family regex parser",
			ValidationCommand: validationCommand(scanner.LangJavaScript),
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return languageSortRank(infos[i].Language) < languageSortRank(infos[j].Language)
	})
	return infos
}

func outputLanguages(resp languagesResponse) error {
	fmt.Printf("\n=== TestGen Supported Languages ===\n\n")
	fmt.Printf("Total language families: %d\n\n", resp.Count)
	for _, lang := range resp.Languages {
		fmt.Printf("- %s\n", displayLanguageName(lang.Language))
		fmt.Printf("  Extensions: %s\n", strings.Join(lang.Extensions, ", "))
		fmt.Printf("  Default framework: %s\n", lang.DefaultFramework)
		fmt.Printf("  Supported frameworks: %s\n", strings.Join(lang.Frameworks, ", "))
		fmt.Printf("  Adapter: %s (%s)\n", lang.Adapter, lang.Parser)
		if len(lang.Aliases) > 0 {
			fmt.Printf("  Aliases: %s\n", strings.Join(lang.Aliases, ", "))
		}
		if lang.ValidationCommand != "" {
			fmt.Printf("  Typical validation: %s\n", lang.ValidationCommand)
		}
	}
	fmt.Println()
	return nil
}

func languageAliases(lang string) []string {
	switch lang {
	case scanner.LangGo:
		return []string{"golang"}
	case scanner.LangPython:
		return []string{"py", "python3"}
	case scanner.LangJavaScript:
		return []string{"js", "node", "nodejs"}
	case scanner.LangTypeScript:
		return []string{"ts"}
	case scanner.LangJava:
		return []string{"jdk", "openjdk", "jvm"}
	case scanner.LangCSharp:
		return []string{"cs", "c#", "dotnet", ".net"}
	case scanner.LangPHP:
		return []string{"php8", "php7"}
	case scanner.LangRuby:
		return []string{"rb", "rails"}
	case scanner.LangCPP:
		return []string{"c++", "cc", "cxx", "cplusplus"}
	case scanner.LangKotlin:
		return []string{"kt", "kts"}
	default:
		return nil
	}
}

func adapterKind(lang string) string {
	if lang == scanner.LangTypeScript {
		return "javascript"
	}
	return lang
}

func parserKind(lang string) string {
	switch lang {
	case scanner.LangGo:
		return "go/parser AST"
	case scanner.LangPython, scanner.LangJavaScript, scanner.LangRust, scanner.LangJava:
		return "language-aware parser"
	default:
		return "regex parser"
	}
}

func validationCommand(lang string) string {
	switch lang {
	case scanner.LangGo:
		return "go test ./..."
	case scanner.LangPython:
		return "pytest"
	case scanner.LangJavaScript, scanner.LangTypeScript:
		return "npm test"
	case scanner.LangRust:
		return "cargo test"
	case scanner.LangJava, scanner.LangKotlin:
		return "mvn test or gradle test"
	case scanner.LangCSharp:
		return "dotnet test"
	case scanner.LangPHP:
		return "vendor/bin/phpunit or vendor/bin/pest"
	case scanner.LangRuby:
		return "bundle exec rspec"
	case scanner.LangCPP:
		return "ctest or configured build runner"
	default:
		return ""
	}
}

func displayLanguageName(lang string) string {
	switch lang {
	case scanner.LangCSharp:
		return "C#"
	case scanner.LangCPP:
		return "C++"
	case scanner.LangPHP:
		return "PHP"
	case scanner.LangJavaScript:
		return "JavaScript"
	case scanner.LangTypeScript:
		return "TypeScript"
	case scanner.LangGo:
		return "Go"
	default:
		if lang == "" {
			return ""
		}
		return strings.ToUpper(lang[:1]) + lang[1:]
	}
}

func languageSortRank(lang string) int {
	order := []string{
		scanner.LangJavaScript,
		scanner.LangTypeScript,
		scanner.LangPython,
		scanner.LangGo,
		scanner.LangRust,
		scanner.LangJava,
		scanner.LangCSharp,
		scanner.LangPHP,
		scanner.LangRuby,
		scanner.LangCPP,
		scanner.LangKotlin,
	}
	for idx, candidate := range order {
		if lang == candidate {
			return idx
		}
	}
	return len(order) + 1
}
