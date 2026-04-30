package adapters

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/princepal9120/testgen-cli/pkg/models"
)

type regexLanguageAdapter struct {
	BaseAdapter
	extensions      []string
	methodPattern   *regexp.Regexp
	functionPattern *regexp.Regexp
	classPattern    *regexp.Regexp
	importPattern   *regexp.Regexp
	testSuffix      string
	testExt         string
	testDir         string
	runCommand      []string
	promptName      string
	promptHints     []string
}

func (a *regexLanguageAdapter) CanHandle(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	for _, candidate := range a.extensions {
		if ext == candidate {
			return true
		}
	}
	return false
}

func (a *regexLanguageAdapter) ParseFile(content string) (*models.AST, error) {
	ast := &models.AST{
		Language:    a.language,
		Definitions: make([]*models.Definition, 0),
		Imports:     make([]string, 0),
	}
	lines := strings.Split(content, "\n")

	if a.importPattern != nil {
		for _, line := range lines {
			if match := a.importPattern.FindStringSubmatch(strings.TrimSpace(line)); len(match) > 1 {
				ast.Imports = append(ast.Imports, strings.TrimSpace(match[1]))
			}
		}
	}

	var currentClass string
	if a.classPattern != nil {
		for _, line := range lines {
			if match := a.classPattern.FindStringSubmatch(line); len(match) > 1 {
				currentClass = strings.TrimSpace(match[1])
				break
			}
		}
	}

	seen := map[string]bool{}
	for i, line := range lines {
		for _, candidate := range []struct {
			pattern  *regexp.Regexp
			isMethod bool
		}{
			{a.methodPattern, true},
			{a.functionPattern, false},
		} {
			if candidate.pattern == nil {
				continue
			}
			match := candidate.pattern.FindStringSubmatch(line)
			if len(match) < 2 {
				continue
			}

			name := strings.TrimSpace(match[len(match)-2])
			params := strings.TrimSpace(match[len(match)-1])
			returnType := ""
			if a.language == "kotlin" && len(match) >= 4 {
				name = strings.TrimSpace(match[1])
				params = strings.TrimSpace(match[2])
				returnType = strings.TrimSpace(match[3])
			}
			trimmedLine := strings.TrimSpace(line)
			if name == "" || strings.HasPrefix(trimmedLine, "throw ") || isIgnoredFunctionName(name) || seen[fmt.Sprintf("%d:%s", i, name)] {
				continue
			}
			seen[fmt.Sprintf("%d:%s", i, name)] = true

			if returnType == "" && len(match) >= 4 {
				returnType = strings.TrimSpace(match[1])
			}
			endLine := findBraceBlockEnd(lines, i)
			body := ""
			if endLine > i+1 {
				body = strings.Join(lines[i+1:endLine], "\n")
			}

			ast.Definitions = append(ast.Definitions, &models.Definition{
				Name:       name,
				Signature:  strings.TrimSpace(line),
				Body:       body,
				StartLine:  i + 1,
				EndLine:    endLine,
				IsMethod:   candidate.isMethod || currentClass != "",
				ClassName:  currentClass,
				Parameters: parseGenericParams(params),
				ReturnType: returnType,
			})
		}
	}
	return ast, nil
}

func (a *regexLanguageAdapter) ExtractDefinitions(ast *models.AST) ([]*models.Definition, error) {
	if ast == nil {
		return nil, fmt.Errorf("nil AST provided")
	}
	return ast.Definitions, nil
}

func (a *regexLanguageAdapter) SelectFramework(projectPath string) string {
	return a.defaultFW
}

func (a *regexLanguageAdapter) GenerateTestPath(sourcePath string, outputDir string) string {
	ext := a.testExt
	if ext == "" {
		ext = filepath.Ext(sourcePath)
	}
	base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath)) + a.testSuffix + ext
	if outputDir != "" {
		return filepath.Join(outputDir, base)
	}
	if a.testDir != "" {
		return filepath.Join(filepath.Dir(sourcePath), a.testDir, base)
	}
	return filepath.Join(filepath.Dir(sourcePath), base)
}

func (a *regexLanguageAdapter) FormatTestCode(code string) (string, error) {
	return strings.TrimSpace(code) + "\n", nil
}

func (a *regexLanguageAdapter) GetPromptTemplate(testType string) string {
	hints := strings.Join(a.promptHints, "\n- ")
	if hints != "" {
		hints = "\n- " + hints
	}
	base := fmt.Sprintf(`Generate %s tests for this %s source file using %s.

Requirements:
- Cover exported/public functions and meaningful edge cases.
- Match idiomatic %s project conventions.
- Include clear assertions and minimal mocking.
- Return only complete test code.%s`, testType, a.language, a.defaultFW, a.promptName, hints)

	switch testType {
	case "edge-cases":
		return base + "\n- Focus on boundaries, empty inputs, null/nil values, and unusual valid inputs."
	case "negative":
		return base + "\n- Focus on invalid inputs, exceptions/errors, and failure branches."
	case "integration":
		return base + "\n- Prefer realistic integration boundaries with safe fakes for external systems."
	default:
		return base
	}
}

func (a *regexLanguageAdapter) ValidateTests(testCode string, testPath string) error {
	if strings.TrimSpace(testCode) == "" {
		return fmt.Errorf("empty test code")
	}
	return nil
}

func (a *regexLanguageAdapter) RunTests(testDir string) (*models.TestResults, error) {
	if len(a.runCommand) == 0 {
		return &models.TestResults{ExitCode: 0, Output: "No default test command configured"}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.runCommand[0], a.runCommand[1:]...)
	cmd.Dir = testDir
	output, err := cmd.CombinedOutput()
	result := &models.TestResults{Output: string(output)}
	if err != nil {
		result.ExitCode = 1
		return result, err
	}
	return result, nil
}

func parseGenericParams(paramStr string) []models.Param {
	params := []models.Param{}
	for _, part := range splitTopLevel(paramStr, ',') {
		part = strings.TrimSpace(part)
		if part == "" || part == "self" || part == "this" {
			continue
		}
		part = strings.TrimPrefix(part, "...")
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}
		param := models.Param{Name: strings.Trim(fields[len(fields)-1], "*$&?")}
		if len(fields) > 1 {
			param.Type = strings.Join(fields[:len(fields)-1], " ")
		}
		if colon := strings.Index(param.Name, ":"); colon > 0 {
			param.Type = strings.TrimSpace(param.Name[colon+1:])
			param.Name = strings.TrimSpace(param.Name[:colon])
		}
		params = append(params, param)
	}
	return params
}

func splitTopLevel(s string, sep rune) []string {
	parts := []string{}
	var current strings.Builder
	depth := 0
	for _, ch := range s {
		switch ch {
		case '(', '[', '{', '<':
			depth++
		case ')', ']', '}', '>':
			if depth > 0 {
				depth--
			}
		}
		if ch == sep && depth == 0 {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(ch)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func findBraceBlockEnd(lines []string, startIdx int) int {
	braceCount := 0
	foundOpen := false
	for i := startIdx; i < len(lines); i++ {
		for _, ch := range lines[i] {
			if ch == '{' {
				braceCount++
				foundOpen = true
			}
			if ch == '}' {
				braceCount--
				if foundOpen && braceCount <= 0 {
					return i + 1
				}
			}
		}
	}
	if !foundOpen {
		return startIdx + 1
	}
	return len(lines)
}

func isIgnoredFunctionName(name string) bool {
	switch name {
	case "if", "for", "while", "switch", "catch", "return", "function", "class", "interface", "struct":
		return true
	default:
		return false
	}
}
