package adapters

import "regexp"

func NewCSharpAdapter() LanguageAdapter {
	return &regexLanguageAdapter{
		BaseAdapter:   BaseAdapter{language: "csharp", frameworks: []string{"xunit", "nunit", "mstest"}, defaultFW: "xunit"},
		extensions:    []string{".cs"},
		classPattern:  regexp.MustCompile(`\b(?:class|record|struct)\s+(\w+)`),
		importPattern: regexp.MustCompile(`^using\s+([\w.]+)\s*;`),
		methodPattern: regexp.MustCompile(`^\s*(?:public|private|protected|internal)?\s*(?:static\s+)?(?:async\s+)?([\w<>?\[\],\s]+)\s+(\w+)\s*\(([^)]*)\)`),
		testSuffix:    "Tests",
		testExt:       ".cs",
		promptName:    "C#",
		promptHints:   []string{"Use xUnit [Fact] or [Theory] tests by default.", "Use FluentAssertions only if the project already uses it."},
	}
}
