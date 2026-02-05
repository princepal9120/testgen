package generator

import (
	"strings"
	"testing"

	"github.com/princepal9120/testgen-cli/pkg/models"
)

func TestExtractCodeFromResponse_WithLanguageCodeFence(t *testing.T) {
	response := "Some preface\n```go\nfunc TestAdd(t *testing.T) {}\n```\nSome note"
	got := extractCodeFromResponse(response, "go")

	if got != "func TestAdd(t *testing.T) {}" {
		t.Fatalf("unexpected extracted code: %q", got)
	}
}

func TestExtractCodeFromResponse_WithGenericCodeFence(t *testing.T) {
	response := "```\ndef test_add():\n    assert 1 + 1 == 2\n```"
	got := extractCodeFromResponse(response, "python")

	if got != "def test_add():\n    assert 1 + 1 == 2" {
		t.Fatalf("unexpected extracted code: %q", got)
	}
}

func TestExtractCodeFromResponse_WithoutCodeFence(t *testing.T) {
	response := "   fn add(a: i32, b: i32) -> i32 { a + b }   "
	got := extractCodeFromResponse(response, "rust")

	if got != "fn add(a: i32, b: i32) -> i32 { a + b }" {
		t.Fatalf("unexpected extracted code: %q", got)
	}
}

func TestParseStructuredOutput_Success(t *testing.T) {
	response := `metadata before {
		"test_name": "TestAdd",
		"test_code": "def test_add(): assert add(1, 2) == 3",
		"imports": ["pytest"],
		"edge_cases_covered": ["negative"],
		"mocked_dependencies": []
	} metadata after`

	result, err := parseStructuredOutput(response)
	if err != nil {
		t.Fatalf("expected parse success, got error: %v", err)
	}

	if result.TestName != "TestAdd" {
		t.Fatalf("unexpected test_name: %q", result.TestName)
	}
	if !strings.Contains(result.TestCode, "test_add") {
		t.Fatalf("expected test code to contain test name, got: %q", result.TestCode)
	}
	if len(result.Imports) != 1 || result.Imports[0] != "pytest" {
		t.Fatalf("unexpected imports: %#v", result.Imports)
	}
}

func TestParseStructuredOutput_NoJSON(t *testing.T) {
	_, err := parseStructuredOutput("plain text response")
	if err == nil {
		t.Fatal("expected error when JSON is missing")
	}
}

func TestParseStructuredOutput_InvalidJSON(t *testing.T) {
	_, err := parseStructuredOutput("{\"test_name\":")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestPostProcess_GoAddsPackageAndImports(t *testing.T) {
	engine := &Engine{}
	ast := &models.AST{Package: "calculator"}

	code := "func TestAdd(t *testing.T) {}"
	result := engine.postProcess(code, nil, "go", ast)

	if !strings.Contains(result, "package calculator_test") {
		t.Fatalf("expected generated package declaration, got:\n%s", result)
	}
	if !strings.Contains(result, "github.com/stretchr/testify/assert") {
		t.Fatalf("expected testify assert import, got:\n%s", result)
	}
	if !strings.Contains(result, code) {
		t.Fatalf("expected original test code to be present, got:\n%s", result)
	}
}

func TestPostProcess_GoPreservesExistingPackageDeclaration(t *testing.T) {
	engine := &Engine{}
	ast := &models.AST{Package: "calculator"}

	code := "package calculator_test\n\nfunc TestAlreadyPackaged(t *testing.T) {}"
	result := engine.postProcess(code, nil, "go", ast)

	if result != code {
		t.Fatalf("expected code to be unchanged when package exists, got:\n%s", result)
	}
}
