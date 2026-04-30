package adapters

import (
	"path/filepath"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandedLanguageAdapters_ParseDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		adapter  LanguageAdapter
		file     string
		code     string
		expected string
	}{
		{
			name:     "csharp",
			adapter:  NewCSharpAdapter(),
			file:     "Calculator.cs",
			expected: "Add",
			code: `using System;
public class Calculator {
  public int Add(int a, int b) { return a + b; }
}`,
		},
		{
			name:     "php",
			adapter:  NewPHPAdapter(),
			file:     "calculator.php",
			expected: "add",
			code: `<?php
function add(int $a, int $b): int { return $a + $b; }
class Calculator { public function subtract($a, $b) { return $a - $b; } }`,
		},
		{
			name:     "ruby",
			adapter:  NewRubyAdapter(),
			file:     "calculator.rb",
			expected: "add",
			code: `class Calculator
  def add(a, b)
    a + b
  end
end`,
		},
		{
			name:     "cpp",
			adapter:  NewCPPAdapter(),
			file:     "calculator.cpp",
			expected: "add",
			code: `#include <stdexcept>
int add(int a, int b) { return a + b; }`,
		},
		{
			name:     "kotlin",
			adapter:  NewKotlinAdapter(),
			file:     "Calculator.kt",
			expected: "add",
			code: `package demo
class Calculator {
  fun add(a: Int, b: Int): Int { return a + b }
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.adapter.CanHandle(tt.file))
			ast, err := tt.adapter.ParseFile(tt.code)
			require.NoError(t, err)
			require.NotEmpty(t, ast.Definitions)
			assert.Equal(t, tt.expected, ast.Definitions[0].Name)
			assert.NotEmpty(t, tt.adapter.GetPromptTemplate("unit"))
			assert.NotEmpty(t, tt.adapter.GenerateTestPath(filepath.Join("src", tt.file), ""))
		})
	}
}

func TestRegistryIncludesExpandedLanguages(t *testing.T) {
	registry := DefaultRegistry()
	for _, lang := range []string{scanner.LangCSharp, scanner.LangPHP, scanner.LangRuby, scanner.LangCPP, scanner.LangKotlin} {
		assert.True(t, registry.HasAdapter(lang), "expected adapter for %s", lang)
	}
}

func TestScannerDetectsExpandedLanguages(t *testing.T) {
	tests := map[string]string{
		"service.cs":     scanner.LangCSharp,
		"index.php":      scanner.LangPHP,
		"calculator.rb":  scanner.LangRuby,
		"calculator.cpp": scanner.LangCPP,
		"Calculator.kt":  scanner.LangKotlin,
		"Calculator.kts": scanner.LangKotlin,
		"calculator.cxx": scanner.LangCPP,
		"calculator.hpp": scanner.LangCPP,
	}
	for file, expected := range tests {
		assert.Equal(t, expected, scanner.DetectLanguage(file))
	}
}
