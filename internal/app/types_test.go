package app

import (
	"testing"

	"github.com/princepal9120/testgen-cli/internal/llm"
)

func TestNormalizeGenerateRequestRejectsUnknownVersion(t *testing.T) {
	t.Parallel()

	req := GenerateRequest{APIVersion: "v999"}
	err := normalizeGenerateRequest(&req)
	if err == nil {
		t.Fatal("expected unsupported version error")
	}
	if classifyFailure(err) != FailureCodeInvalidRequest {
		t.Fatalf("expected invalid_request failure code, got %q", classifyFailure(err))
	}
}

func TestGenerateRequestResolvedWriteSemantics(t *testing.T) {
	t.Parallel()

	if !GenerateRequest{}.ResolvedWriteFiles() {
		t.Fatal("legacy request without write_files should still default to writes")
	}

	writeFiles := false
	req := GenerateRequest{WriteFiles: &writeFiles}
	if req.ResolvedWriteFiles() {
		t.Fatal("explicit write_files=false should force dry-run semantics")
	}
	if !req.ResolvedDryRun() {
		t.Fatal("expected resolved dry-run when write_files=false")
	}

	writeFiles = true
	req = GenerateRequest{DryRun: true, WriteFiles: &writeFiles}
	if req.ResolvedDryRun() {
		t.Fatal("write_files=true should override dry_run request")
	}
}

func TestNewGenerateFailureResponseIncludesFailureCode(t *testing.T) {
	t.Parallel()

	req := GenerateRequest{RequestID: "req_123"}
	resp := NewGenerateFailureResponse(req, llm.ErrNoAPIKey, "/tmp/example.py")

	if resp.APIVersion != APIVersion {
		t.Fatalf("expected api version %q, got %q", APIVersion, resp.APIVersion)
	}
	if resp.RequestID != req.RequestID {
		t.Fatalf("expected request id %q, got %q", req.RequestID, resp.RequestID)
	}
	if resp.FailureCode != FailureCodeMissingAPIKey {
		t.Fatalf("expected missing_api_key, got %q", resp.FailureCode)
	}
	if resp.Success {
		t.Fatal("expected failure response")
	}
}
