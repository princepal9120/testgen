package adapters

import "regexp"

func NewKotlinAdapter() LanguageAdapter {
	return &regexLanguageAdapter{
		BaseAdapter:     BaseAdapter{language: "kotlin", frameworks: []string{"junit5", "kotest", "mockk"}, defaultFW: "junit5"},
		extensions:      []string{".kt", ".kts"},
		classPattern:    regexp.MustCompile(`\b(?:class|object|data\s+class)\s+(\w+)`),
		importPattern:   regexp.MustCompile(`^import\s+([\w.*]+)`),
		methodPattern:   regexp.MustCompile(`^\s*(?:public|private|protected|internal)?\s*(?:suspend\s+)?fun\s+(?:\w+\.)?(\w+)\s*\(([^)]*)\)\s*(?::\s*([\w?<>,.]+))?`),
		functionPattern: regexp.MustCompile(`^\s*(?:public|private|protected|internal)?\s*(?:suspend\s+)?fun\s+(\w+)\s*\(([^)]*)\)\s*(?::\s*([\w?<>,.]+))?`),
		testSuffix:      "Test",
		testExt:         ".kt",
		promptName:      "Kotlin",
		promptHints:     []string{"Use JUnit 5 by default.", "Use Kotest or MockK only when already present in the project."},
	}
}
