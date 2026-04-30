package adapters

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/princepal9120/testgen-cli/pkg/models"
)

type RubyAdapter struct{ *regexLanguageAdapter }

func NewRubyAdapter() LanguageAdapter {
	return &RubyAdapter{&regexLanguageAdapter{
		BaseAdapter:     BaseAdapter{language: "ruby", frameworks: []string{"rspec", "minitest"}, defaultFW: "rspec"},
		extensions:      []string{".rb"},
		classPattern:    regexp.MustCompile(`^\s*class\s+(\w+)`),
		importPattern:   regexp.MustCompile(`^\s*require(?:_relative)?\s+['"]([^'"]+)['"]`),
		functionPattern: regexp.MustCompile(`^\s*def\s+(?:self\.)?(\w+[!?=]?)\s*(?:\(([^)]*)\)|\s*([^#]*))?`),
		testSuffix:      "_spec",
		testExt:         ".rb",
		testDir:         "spec",
		promptName:      "Ruby",
		promptHints:     []string{"Use RSpec describe/context/it blocks by default.", "Use Minitest only when the repo already uses it."},
	}}
}

func (a *RubyAdapter) ParseFile(content string) (*models.AST, error) {
	ast := &models.AST{Language: "ruby", Definitions: []*models.Definition{}, Imports: []string{}}
	lines := strings.Split(content, "\n")
	var currentClass string
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if match := a.importPattern.FindStringSubmatch(trimmed); match != nil {
			ast.Imports = append(ast.Imports, match[1])
		}
		if match := a.classPattern.FindStringSubmatch(line); match != nil {
			currentClass = match[1]
			continue
		}
		if match := a.functionPattern.FindStringSubmatch(line); match != nil {
			name := match[1]
			params := ""
			if len(match) > 2 && match[2] != "" {
				params = match[2]
			} else if len(match) > 3 {
				params = strings.TrimSpace(match[3])
			}
			ast.Definitions = append(ast.Definitions, &models.Definition{
				Name:       name,
				Signature:  strings.TrimSpace(line),
				StartLine:  i + 1,
				EndLine:    findRubyDefEnd(lines, i),
				IsMethod:   currentClass != "",
				ClassName:  currentClass,
				Parameters: parseGenericParams(params),
			})
		}
	}
	return ast, nil
}

func (a *RubyAdapter) GenerateTestPath(sourcePath string, outputDir string) string {
	base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath)) + "_spec.rb"
	if outputDir != "" {
		return filepath.Join(outputDir, base)
	}
	return filepath.Join(filepath.Dir(sourcePath), "spec", base)
}

func findRubyDefEnd(lines []string, startIdx int) int {
	depth := 0
	for i := startIdx; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") || strings.HasPrefix(trimmed, "module ") || strings.HasPrefix(trimmed, "do") || strings.HasSuffix(trimmed, " do") {
			depth++
		}
		if trimmed == "end" {
			depth--
			if depth <= 0 {
				return i + 1
			}
		}
	}
	return len(lines)
}

func (a *RubyAdapter) ExtractDefinitions(ast *models.AST) ([]*models.Definition, error) {
	if ast == nil {
		return nil, fmt.Errorf("nil AST provided")
	}
	return ast.Definitions, nil
}
