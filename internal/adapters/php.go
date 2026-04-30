package adapters

import "regexp"

func NewPHPAdapter() LanguageAdapter {
	return &regexLanguageAdapter{
		BaseAdapter:     BaseAdapter{language: "php", frameworks: []string{"phpunit", "pest"}, defaultFW: "phpunit"},
		extensions:      []string{".php"},
		classPattern:    regexp.MustCompile(`\bclass\s+(\w+)`),
		importPattern:   regexp.MustCompile(`^use\s+([^;]+);`),
		methodPattern:   regexp.MustCompile(`^\s*(?:public|private|protected)?\s*(?:static\s+)?function\s+(\w+)\s*\(([^)]*)\)`),
		functionPattern: regexp.MustCompile(`^\s*function\s+(\w+)\s*\(([^)]*)\)`),
		testSuffix:      "Test",
		testExt:         ".php",
		promptName:      "PHP",
		promptHints:     []string{"Use PHPUnit by default.", "Prefer strict assertions and data providers when useful."},
	}
}
