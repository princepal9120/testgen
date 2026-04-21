package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/princepal9120/testgen-cli/pkg/models"
)

func TestValidatorFindsPythonTestsInSiblingTestsDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	testsDir := filepath.Join(root, "tests")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.MkdirAll(testsDir, 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}

	sourcePath := filepath.Join(srcDir, "math.py")
	testPath := filepath.Join(testsDir, "test_math.py")
	if err := os.WriteFile(sourcePath, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(testPath, []byte("def test_add():\n    assert True\n"), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}

	validator := NewValidator(Config{})
	result, err := validator.Validate(srcDir, []*models.SourceFile{{
		Path:     sourcePath,
		Language: "python",
	}})
	if err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
	if result.FilesWithTests != 1 {
		t.Fatalf("expected 1 file with tests, got %d", result.FilesWithTests)
	}
	if len(result.FilesMissingTests) != 0 {
		t.Fatalf("expected no missing tests, got %d", len(result.FilesMissingTests))
	}
}

func TestValidatorCountsInlineRustTestsAsCovered(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourcePath := filepath.Join(root, "lib.rs")
	if err := os.WriteFile(sourcePath, []byte("pub fn add(a:i32,b:i32)->i32 { a+b }\n#[cfg(test)]\nmod tests {\n #[test]\n fn works() { assert_eq!(2, 1+1); }\n}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	validator := NewValidator(Config{})
	result, err := validator.Validate(root, []*models.SourceFile{{
		Path:     sourcePath,
		Language: "rust",
	}})
	if err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
	if result.FilesWithTests != 1 {
		t.Fatalf("expected inline Rust tests to count as coverage, got %d", result.FilesWithTests)
	}
	if len(result.FilesMissingTests) != 0 {
		t.Fatalf("expected no missing tests, got %d", len(result.FilesMissingTests))
	}
}

func TestValidatorRejectsUnrecognizableDiscoveredTestFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	testsDir := filepath.Join(root, "tests")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.MkdirAll(testsDir, 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}

	sourcePath := filepath.Join(srcDir, "math.py")
	testPath := filepath.Join(testsDir, "test_math.py")
	if err := os.WriteFile(sourcePath, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(testPath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}

	validator := NewValidator(Config{})
	result, err := validator.Validate(srcDir, []*models.SourceFile{{
		Path:     sourcePath,
		Language: "python",
	}})
	if err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
	if result.FilesWithTests != 0 {
		t.Fatalf("expected invalid discovered tests not to count, got %d", result.FilesWithTests)
	}
	if len(result.FilesMissingTests) != 1 {
		t.Fatalf("expected 1 missing test file, got %d", len(result.FilesMissingTests))
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected recognizable validation error for bogus discovered test file")
	}
}
