package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRustAdapter_ParseFile(t *testing.T) {
	adapter := NewRustAdapter()

	t.Run("Parse basic function", func(t *testing.T) {
		code := `
fn add(a: i32, b: i32) -> i32 {
    a + b
}
`
		ast, err := adapter.ParseFile(code)
		assert.NoError(t, err)
		assert.Len(t, ast.Definitions, 1)
		assert.Equal(t, "add", ast.Definitions[0].Name)
	})

	t.Run("Parse public async function", func(t *testing.T) {
		code := `
pub async fn fetch_data() -> Result<String, Error> {
    Ok("data".to_string())
}
`
		ast, err := adapter.ParseFile(code)
		assert.NoError(t, err)
		assert.Len(t, ast.Definitions, 1)

		def := ast.Definitions[0]
		assert.Equal(t, "fetch_data", def.Name)
		assert.Contains(t, def.Signature, "async")
		assert.Contains(t, def.Signature, "pub")
	})

	t.Run("Parse impl method", func(t *testing.T) {
		code := `
impl User {
    pub fn new(name: String) -> Self {
        User { name }
    }
}
`
		ast, err := adapter.ParseFile(code)
		assert.NoError(t, err)
		assert.Len(t, ast.Definitions, 1)

		def := ast.Definitions[0]
		assert.Equal(t, "new", def.Name)
		assert.True(t, def.IsMethod)
		assert.Equal(t, "User", def.ClassName)
	})
}

func TestRustAdapter_GetPromptTemplate(t *testing.T) {
	adapter := NewRustAdapter()

	prompt := adapter.GetPromptTemplate("unit")
	assert.Contains(t, prompt, "idiomatic Rust tests")
	assert.Contains(t, prompt, "#[cfg(test)]")
}

func TestRustAdapter_GenerateTestPath(t *testing.T) {
	adapter := NewRustAdapter()

	// Inline tests (default behavior if no tests dir)
	path := adapter.GenerateTestPath("/src/lib.rs", "")
	assert.Contains(t, filepath.ToSlash(path), "lib.rs.test") // This is our fallback

	// Explicit output dir
	pathWithDir := adapter.GenerateTestPath("/src/lib.rs", "/tests")
	assert.Equal(t, "/tests/lib_test.rs", filepath.ToSlash(pathWithDir))
}

func TestFindCargoProjectRoot(t *testing.T) {
	t.Run("finds parent cargo project", func(t *testing.T) {
		root := t.TempDir()
		nested := filepath.Join(root, "src", "module")
		assert.NoError(t, os.MkdirAll(nested, 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(root, "Cargo.toml"), []byte("[package]\nname = \"demo\"\nversion = \"0.1.0\"\n"), 0o644))

		cargoRoot, found := findCargoProjectRoot(nested)
		assert.True(t, found)
		assert.Equal(t, root, cargoRoot)
	})

	t.Run("stops at filesystem root when cargo project missing", func(t *testing.T) {
		root := t.TempDir()
		nested := filepath.Join(root, "src", "module")
		assert.NoError(t, os.MkdirAll(nested, 0o755))

		cargoRoot, found := findCargoProjectRoot(nested)
		assert.False(t, found)
		assert.Empty(t, cargoRoot)
	})
}

func TestRustAdapter_RunTestsSkipsWhenCargoProjectMissing(t *testing.T) {
	adapter := NewRustAdapter()
	root := t.TempDir()

	results, err := adapter.RunTests(root)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, results.ExitCode)
	assert.Contains(t, results.Output, "skipped cargo test")
}
