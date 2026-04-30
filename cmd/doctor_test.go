package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/app"
)

func TestBuildDoctorResponseDetectsRepoReadiness(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "package.json"), `{"devDependencies":{"vitest":"latest"},"scripts":{"test":"vitest"}}`)
	writeTestFile(t, filepath.Join(dir, "src", "math.ts"), "export function add(a: number, b: number) { return a + b }\n")
	writeTestFile(t, filepath.Join(dir, "tests", "math.test.ts"), "")
	t.Setenv("OPENAI_API_KEY", "test-key")

	resp, err := buildDoctorResponse(dir, true)
	if err != nil {
		t.Fatalf("build doctor response: %v", err)
	}
	if resp.APIVersion != app.APIVersion {
		t.Fatalf("expected API version %q, got %q", app.APIVersion, resp.APIVersion)
	}
	if !resp.Success {
		t.Fatal("expected success response")
	}
	if resp.SourceFileCount != 1 {
		t.Fatalf("expected one source file, got %d", resp.SourceFileCount)
	}
	if !includes(resp.DetectedLanguages, "typescript") {
		t.Fatalf("expected typescript in detected languages, got %#v", resp.DetectedLanguages)
	}
	if !includes(resp.DetectedFrameworks["typescript"], "vitest") {
		t.Fatalf("expected vitest framework, got %#v", resp.DetectedFrameworks)
	}
	if !includes(resp.ExistingTestDirs, "tests") {
		t.Fatalf("expected tests dir, got %#v", resp.ExistingTestDirs)
	}
	if !includes(resp.NativeTestCommands, "npm test") {
		t.Fatalf("expected npm test candidate, got %#v", resp.NativeTestCommands)
	}
	if !anyProviderConfigured(resp.ProviderKeys) {
		t.Fatalf("expected configured provider key, got %#v", resp.ProviderKeys)
	}
	if resp.SuggestedCommand == "" {
		t.Fatal("expected suggested command")
	}

	encoded, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal doctor response: %v", err)
	}
	var decoded doctorResponse
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode doctor response: %v", err)
	}
}

func TestBuildDoctorResponseWarnsWhenNoSourcesOrKeys(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GROQ_API_KEY", "")

	resp, err := buildDoctorResponse(dir, true)
	if err != nil {
		t.Fatalf("build doctor response: %v", err)
	}
	if resp.SourceFileCount != 0 {
		t.Fatalf("expected no source files, got %d", resp.SourceFileCount)
	}
	if len(resp.Warnings) < 2 {
		t.Fatalf("expected source and provider warnings, got %#v", resp.Warnings)
	}
	if resp.SuggestedCommand != "testgen languages --output-format=json" {
		t.Fatalf("unexpected suggested command: %q", resp.SuggestedCommand)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
