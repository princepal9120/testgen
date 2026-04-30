package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/princepal9120/testgen-cli/internal/scanner"
	"github.com/spf13/cobra"
)

type providerStatus struct {
	Provider      string   `json:"provider"`
	Configured    bool     `json:"configured"`
	EnvVars       []string `json:"env_vars"`
	ConfiguredEnv string   `json:"configured_env,omitempty"`
}

type doctorResponse struct {
	APIVersion             string              `json:"api_version"`
	Success                bool                `json:"success"`
	Path                   string              `json:"path"`
	DetectedLanguages      []string            `json:"detected_languages"`
	DetectedFrameworks     map[string][]string `json:"detected_frameworks"`
	SourceFileCount        int                 `json:"source_file_count"`
	ExistingTestDirs       []string            `json:"existing_test_dirs"`
	NativeTestCommands     []string            `json:"native_test_commands"`
	ProviderKeys           []providerStatus    `json:"provider_keys"`
	IgnoredDirectories     []string            `json:"ignored_directories"`
	UnsupportedFileSamples []string            `json:"unsupported_file_samples,omitempty"`
	Warnings               []string            `json:"warnings,omitempty"`
	SuggestedCommand       string              `json:"suggested_safe_first_command"`
}

var (
	doctorPath         string
	doctorRecursive    bool
	doctorOutputFormat string
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check whether a repository is ready for TestGen",
	Long: `Scan a repository for supported languages, framework markers, existing tests,
provider API key availability, unsupported files, and a safe first command.`,
	Example: `  testgen doctor --path=.
  testgen doctor --path=. --output-format=json`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().StringVarP(&doctorPath, "path", "p", ".", "repository path to inspect")
	doctorCmd.Flags().BoolVarP(&doctorRecursive, "recursive", "r", true, "scan recursively")
	doctorCmd.Flags().StringVar(&doctorOutputFormat, "output-format", "text", "output format: text, json")
}

func runDoctor(cmd *cobra.Command, args []string) error {
	machineMode := strings.EqualFold(doctorOutputFormat, "json")
	if machineMode {
		previousQuiet := quiet
		quiet = true
		defer func() { quiet = previousQuiet }()
		initLogger()
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}

	resp, err := buildDoctorResponse(doctorPath, doctorRecursive)
	if machineMode {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if resp != nil {
			_ = encoder.Encode(resp)
		}
	}
	if err != nil {
		return err
	}
	if machineMode {
		return nil
	}
	return outputDoctor(resp)
}

func buildDoctorResponse(path string, recursive bool) (*doctorResponse, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return &doctorResponse{APIVersion: app.APIVersion, Success: false, Path: path, Warnings: []string{err.Error()}}, err
	}
	root := absPath
	if !info.IsDir() {
		root = filepath.Dir(absPath)
	}

	s := scanner.New(scanner.Options{Recursive: recursive, IgnoreFile: filepath.Join(root, ".testgenignore")})
	sourceFiles, err := s.Scan(absPath)
	if err != nil {
		return &doctorResponse{APIVersion: app.APIVersion, Success: false, Path: path, Warnings: []string{err.Error()}}, err
	}

	languagesSet := map[string]bool{}
	for _, f := range sourceFiles {
		languagesSet[f.Language] = true
	}
	languages := sortedKeys(languagesSet)
	frameworks := detectFrameworks(root, languagesSet)
	testDirs := findExistingTestDirs(root, recursive)
	commands := nativeTestCommands(root, languagesSet, frameworks)
	unsupported := unsupportedFileSamples(root, recursive, 10)
	providers := providerKeyStatuses()

	warnings := []string{}
	if len(sourceFiles) == 0 {
		warnings = append(warnings, "No supported source files were found.")
	}
	if !anyProviderConfigured(providers) {
		warnings = append(warnings, "No provider API keys were found. Set ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY, or GROQ_API_KEY before generation.")
	}

	return &doctorResponse{
		APIVersion:             app.APIVersion,
		Success:                true,
		Path:                   path,
		DetectedLanguages:      languages,
		DetectedFrameworks:     frameworks,
		SourceFileCount:        len(sourceFiles),
		ExistingTestDirs:       testDirs,
		NativeTestCommands:     commands,
		ProviderKeys:           providers,
		IgnoredDirectories:     defaultIgnoredDirectories(),
		UnsupportedFileSamples: unsupported,
		Warnings:               warnings,
		SuggestedCommand:       suggestedDoctorCommand(path, len(sourceFiles) > 0),
	}, nil
}

func outputDoctor(resp *doctorResponse) error {
	fmt.Printf("\n=== TestGen Doctor ===\n\n")
	fmt.Printf("Path: %s\n", resp.Path)
	fmt.Printf("Source files: %d\n", resp.SourceFileCount)
	fmt.Printf("Languages: %s\n", strings.Join(resp.DetectedLanguages, ", "))
	if len(resp.NativeTestCommands) > 0 {
		fmt.Printf("Native test commands: %s\n", strings.Join(resp.NativeTestCommands, "; "))
	}
	if len(resp.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range resp.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
	fmt.Printf("Suggested first command: %s\n\n", resp.SuggestedCommand)
	return nil
}

func detectFrameworks(root string, languages map[string]bool) map[string][]string {
	frameworks := map[string][]string{}
	if languages[scanner.LangJavaScript] || languages[scanner.LangTypeScript] {
		frameworks[scanner.LangJavaScript] = detectJSFrameworks(root)
		if languages[scanner.LangTypeScript] {
			frameworks[scanner.LangTypeScript] = frameworks[scanner.LangJavaScript]
		}
	}
	if languages[scanner.LangPython] {
		frameworks[scanner.LangPython] = detectPythonFrameworks(root)
	}
	if languages[scanner.LangGo] {
		frameworks[scanner.LangGo] = []string{"testing"}
	}
	if languages[scanner.LangRust] {
		frameworks[scanner.LangRust] = []string{"cargo-test"}
	}
	if languages[scanner.LangJava] {
		frameworks[scanner.LangJava] = detectJavaFrameworks(root)
	}
	if languages[scanner.LangCSharp] {
		frameworks[scanner.LangCSharp] = []string{"dotnet-test"}
	}
	if languages[scanner.LangPHP] {
		frameworks[scanner.LangPHP] = detectPHPFrameworks(root)
	}
	if languages[scanner.LangRuby] {
		frameworks[scanner.LangRuby] = detectRubyFrameworks(root)
	}
	if languages[scanner.LangCPP] {
		frameworks[scanner.LangCPP] = []string{"ctest"}
	}
	if languages[scanner.LangKotlin] {
		frameworks[scanner.LangKotlin] = detectJavaFrameworks(root)
	}
	return frameworks
}

func detectJSFrameworks(root string) []string {
	markers := readLower(filepath.Join(root, "package.json"))
	return uniqueOrDefault([]string{
		containsFramework(markers, "vitest", "vitest"),
		containsFramework(markers, "jest", "jest"),
		containsFramework(markers, "mocha", "mocha"),
		containsFramework(markers, "playwright", "playwright"),
	}, []string{"jest"})
}

func detectPythonFrameworks(root string) []string {
	content := strings.Join([]string{readLower(filepath.Join(root, "pyproject.toml")), readLower(filepath.Join(root, "requirements.txt")), readLower(filepath.Join(root, "setup.cfg"))}, "\n")
	return uniqueOrDefault([]string{containsFramework(content, "pytest", "pytest"), containsFramework(content, "unittest", "unittest")}, []string{"pytest"})
}

func detectJavaFrameworks(root string) []string {
	content := strings.Join([]string{readLower(filepath.Join(root, "pom.xml")), readLower(filepath.Join(root, "build.gradle")), readLower(filepath.Join(root, "build.gradle.kts"))}, "\n")
	return uniqueOrDefault([]string{containsFramework(content, "junit", "junit"), containsFramework(content, "testng", "testng"), containsFramework(content, "kotest", "kotest"), containsFramework(content, "mockk", "mockk")}, []string{"junit"})
}

func detectPHPFrameworks(root string) []string {
	content := readLower(filepath.Join(root, "composer.json"))
	return uniqueOrDefault([]string{containsFramework(content, "phpunit", "phpunit"), containsFramework(content, "pestphp", "pest")}, []string{"phpunit"})
}

func detectRubyFrameworks(root string) []string {
	content := strings.Join([]string{readLower(filepath.Join(root, "Gemfile")), readLower(filepath.Join(root, ".rspec"))}, "\n")
	return uniqueOrDefault([]string{containsFramework(content, "rspec", "rspec"), containsFramework(content, "minitest", "minitest")}, []string{"rspec"})
}

func containsFramework(content, needle, framework string) string {
	if strings.Contains(content, needle) {
		return framework
	}
	return ""
}

func uniqueOrDefault(values, fallback []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func readLower(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.ToLower(string(data))
}

func findExistingTestDirs(root string, recursive bool) []string {
	candidates := []string{}
	walkRepo(root, recursive, func(path string, info os.FileInfo) {
		if !info.IsDir() {
			return
		}
		base := strings.ToLower(info.Name())
		if base == "test" || base == "tests" || base == "__tests__" || base == "spec" {
			candidates = append(candidates, relPath(root, path))
		}
	})
	sort.Strings(candidates)
	return candidates
}

func nativeTestCommands(root string, languages map[string]bool, frameworks map[string][]string) []string {
	commands := []string{}
	if languages[scanner.LangJavaScript] || languages[scanner.LangTypeScript] {
		if exists(filepath.Join(root, "package.json")) {
			commands = append(commands, "npm test")
		}
	}
	if languages[scanner.LangPython] {
		commands = append(commands, "pytest")
	}
	if languages[scanner.LangGo] {
		commands = append(commands, "go test ./...")
	}
	if languages[scanner.LangRust] {
		commands = append(commands, "cargo test")
	}
	if languages[scanner.LangJava] || languages[scanner.LangKotlin] {
		if exists(filepath.Join(root, "pom.xml")) {
			commands = append(commands, "mvn test")
		}
		if exists(filepath.Join(root, "build.gradle")) || exists(filepath.Join(root, "build.gradle.kts")) {
			commands = append(commands, "gradle test")
		}
	}
	if languages[scanner.LangCSharp] {
		commands = append(commands, "dotnet test")
	}
	if languages[scanner.LangPHP] {
		if includes(frameworks[scanner.LangPHP], "pest") {
			commands = append(commands, "vendor/bin/pest")
		} else {
			commands = append(commands, "vendor/bin/phpunit")
		}
	}
	if languages[scanner.LangRuby] {
		commands = append(commands, "bundle exec rspec")
	}
	if languages[scanner.LangCPP] {
		commands = append(commands, "ctest")
	}
	return commands
}

func unsupportedFileSamples(root string, recursive bool, limit int) []string {
	supported := map[string]bool{}
	for _, ext := range scanner.GetSupportedExtensions() {
		supported[ext] = true
	}
	ignored := map[string]bool{".md": true, ".txt": true, ".json": true, ".yaml": true, ".yml": true, ".toml": true, ".lock": true, ".sum": true, ".mod": true}
	samples := []string{}
	walkRepo(root, recursive, func(path string, info os.FileInfo) {
		if info.IsDir() || len(samples) >= limit {
			return
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != "" && !supported[ext] && !ignored[ext] {
			samples = append(samples, relPath(root, path))
		}
	})
	return samples
}

func providerKeyStatuses() []providerStatus {
	providers := []providerStatus{
		{Provider: "anthropic", EnvVars: []string{"ANTHROPIC_API_KEY"}},
		{Provider: "openai", EnvVars: []string{"OPENAI_API_KEY"}},
		{Provider: "gemini", EnvVars: []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"}},
		{Provider: "groq", EnvVars: []string{"GROQ_API_KEY"}},
	}
	for i := range providers {
		for _, env := range providers[i].EnvVars {
			if os.Getenv(env) != "" {
				providers[i].Configured = true
				providers[i].ConfiguredEnv = env
				break
			}
		}
	}
	return providers
}

func anyProviderConfigured(providers []providerStatus) bool {
	for _, provider := range providers {
		if provider.Configured {
			return true
		}
	}
	return false
}

func suggestedDoctorCommand(path string, hasSources bool) string {
	if !hasSources {
		return "testgen languages --output-format=json"
	}
	return fmt.Sprintf("testgen generate --path=%s --recursive --dry-run --emit-patch --output-format=json", path)
}

func defaultIgnoredDirectories() []string {
	return []string{"node_modules", "venv", ".venv", "vendor", "target", "__pycache__", ".git", ".idea", ".vscode", "dist", "build", "coverage", ".pytest_cache", ".mypy_cache"}
}

func walkRepo(root string, recursive bool, visit func(string, os.FileInfo)) {
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() && shouldSkipDoctorDir(path, root) {
			return filepath.SkipDir
		}
		if path != root || !info.IsDir() {
			visit(path, info)
		}
		if info.IsDir() && !recursive && path != root {
			return filepath.SkipDir
		}
		return nil
	})
}

func shouldSkipDoctorDir(path, root string) bool {
	if path == root {
		return false
	}
	base := filepath.Base(path)
	for _, ignored := range defaultIgnoredDirectories() {
		if base == ignored {
			return true
		}
	}
	return false
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

func exists(path string) bool { _, err := os.Stat(path); return err == nil }

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func includes(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
