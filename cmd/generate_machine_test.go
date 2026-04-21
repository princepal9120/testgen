package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/app"
	"github.com/spf13/viper"
)

func TestRunGenerateSupportsRequestFileMachineMode(t *testing.T) {
	resetGenerateCommandState()
	viper.Reset()
	logger = nil
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(sourceFile, []byte("# no functions here\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	requestPath := filepath.Join(dir, "request.json")
	request := `{"api_version":"v1","request_id":"req_machine_success","file":"` + sourceFile + `","test_types":["unit"],"dry_run":true}`
	if err := os.WriteFile(requestPath, []byte(request), 0o644); err != nil {
		t.Fatalf("write request file: %v", err)
	}

	genRequestFile = requestPath
	genOutputFormat = "json"

	stdout := captureStdout(t, func() error {
		return runGenerate(generateCmd, nil)
	})

	var resp app.GenerateResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, stdout)
	}
	if resp.APIVersion != app.APIVersion {
		t.Fatalf("expected api version %q, got %q", app.APIVersion, resp.APIVersion)
	}
	if resp.RequestID != "req_machine_success" {
		t.Fatalf("expected request id to round-trip, got %q", resp.RequestID)
	}
	if !resp.DryRun || resp.WriteFiles {
		t.Fatalf("expected dry-run semantics, got dry_run=%v write_files=%v", resp.DryRun, resp.WriteFiles)
	}
}

func TestRunGenerateOutputsStructuredJSONFailure(t *testing.T) {
	resetGenerateCommandState()
	viper.Reset()
	logger = nil

	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GROQ_API_KEY", "")

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(sourceFile, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	requestPath := filepath.Join(dir, "request.json")
	request := `{"api_version":"v1","request_id":"req_machine_failure","file":"` + sourceFile + `","test_types":["unit"],"dry_run":true,"provider":"anthropic"}`
	if err := os.WriteFile(requestPath, []byte(request), 0o644); err != nil {
		t.Fatalf("write request file: %v", err)
	}

	genRequestFile = requestPath
	genOutputFormat = "json"

	var runErr error
	stdout := captureStdout(t, func() error {
		runErr = runGenerate(generateCmd, nil)
		return nil
	})

	if runErr == nil {
		t.Fatal("expected generation failure")
	}

	var resp app.GenerateResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, stdout)
	}
	if resp.Success {
		t.Fatal("expected structured failure response")
	}
	if resp.FailureCode != app.FailureCodeMissingAPIKey {
		t.Fatalf("expected missing_api_key failure code, got %q", resp.FailureCode)
	}
	if resp.Error == "" {
		t.Fatal("expected top-level error message in failure response")
	}
}

func TestRunGenerateTreatsRequestFileAsMachineModeWithoutJSONFlag(t *testing.T) {
	resetGenerateCommandState()
	viper.Reset()
	logger = nil

	t.Setenv("ANTHROPIC_API_KEY", "")

	dir := t.TempDir()
	sourceFile := filepath.Join(dir, "sample.py")
	if err := os.WriteFile(sourceFile, []byte("def add(a, b):\n    return a + b\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	requestPath := filepath.Join(dir, "request.json")
	request := `{"api_version":"v1","request_id":"req_machine_implicit_json","file":"` + sourceFile + `","test_types":["unit"],"dry_run":true,"provider":"anthropic"}`
	if err := os.WriteFile(requestPath, []byte(request), 0o644); err != nil {
		t.Fatalf("write request file: %v", err)
	}

	genRequestFile = requestPath

	var runErr error
	stdout := captureStdoutNoFail(t, func() error {
		runErr = runGenerate(generateCmd, nil)
		return nil
	})

	if runErr == nil {
		t.Fatal("expected structured machine-mode failure")
	}

	var resp app.GenerateResponse
	if err := json.Unmarshal([]byte(stdout), &resp); err != nil {
		t.Fatalf("decode json output: %v\noutput=%s", err, stdout)
	}
	if resp.RequestID != "req_machine_implicit_json" {
		t.Fatalf("expected request id to round-trip, got %q", resp.RequestID)
	}
	if resp.FailureCode != app.FailureCodeMissingAPIKey {
		t.Fatalf("expected missing_api_key failure code, got %q", resp.FailureCode)
	}
}

func resetGenerateCommandState() {
	genPath = ""
	genFile = ""
	genTypes = []string{"unit"}
	genFramework = ""
	genOutput = ""
	genRecursive = false
	genParallel = 2
	genDryRun = false
	genValidate = false
	genOutputFormat = "text"
	genIncludePattern = ""
	genExcludePattern = ""
	genBatchSize = 5
	genReportUsage = false
	genInteractive = false
	genEmitPatch = false
	genRequestFile = ""
	quiet = false
	verbose = false
}

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = writer

	callErr := fn()

	_ = writer.Close()
	os.Stdout = originalStdout

	output, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("read captured stdout: %v", readErr)
	}
	_ = reader.Close()

	if callErr != nil {
		t.Fatalf("unexpected command error: %v", callErr)
	}

	return string(output)
}

func captureStdoutNoFail(t *testing.T, fn func() error) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = writer

	if err := fn(); err != nil {
		t.Fatalf("unexpected command error: %v", err)
	}

	_ = writer.Close()
	os.Stdout = originalStdout

	output, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("read captured stdout: %v", readErr)
	}
	_ = reader.Close()

	return string(output)
}
