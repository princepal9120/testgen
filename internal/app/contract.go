package app

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/princepal9120/testgen-cli/internal/llm"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

const APIVersion = "v1"

// FailureCode provides a stable machine-readable failure taxonomy.
type FailureCode string

const (
	FailureCodeInvalidRequest          FailureCode = "invalid_request"
	FailureCodeNoSourceFiles           FailureCode = "no_source_files"
	FailureCodeUnsupportedLanguage     FailureCode = "unsupported_language"
	FailureCodeMissingAPIKey           FailureCode = "missing_api_key"
	FailureCodeProviderTimeout         FailureCode = "provider_timeout"
	FailureCodeProviderRateLimited     FailureCode = "provider_rate_limited"
	FailureCodeMalformedProviderOutput FailureCode = "malformed_provider_output"
	FailureCodeValidationFailed        FailureCode = "validation_failed"
	FailureCodeWriteFailed             FailureCode = "write_failed"
	FailureCodeScanFailed              FailureCode = "scan_failed"
	FailureCodeCancelled               FailureCode = "cancelled"
	FailureCodeInternalError           FailureCode = "internal_error"
)

var requestCounter atomic.Uint64

func normalizeGenerateRequest(req *GenerateRequest) error {
	if req == nil {
		return fmt.Errorf("generate request is required")
	}
	version, err := normalizeVersion(req.APIVersion)
	if err != nil {
		return err
	}
	req.APIVersion = version
	if req.RequestID == "" {
		req.RequestID = newRequestID()
	}
	if len(req.TestTypes) == 0 {
		req.TestTypes = []string{"unit"}
	}
	return nil
}

func normalizeAnalyzeRequest(req *AnalyzeRequest) error {
	if req == nil {
		return fmt.Errorf("analyze request is required")
	}
	version, err := normalizeVersion(req.APIVersion)
	if err != nil {
		return err
	}
	req.APIVersion = version
	if req.RequestID == "" {
		req.RequestID = newRequestID()
	}
	if req.Detail == "" {
		req.Detail = "summary"
	}
	return nil
}

func normalizeValidateRequest(req *ValidateRequest) error {
	if req == nil {
		return fmt.Errorf("validate request is required")
	}
	version, err := normalizeVersion(req.APIVersion)
	if err != nil {
		return err
	}
	req.APIVersion = version
	if req.RequestID == "" {
		req.RequestID = newRequestID()
	}
	return nil
}

func normalizeVersion(version string) (string, error) {
	if version == "" {
		return APIVersion, nil
	}
	if version != APIVersion {
		return "", fmt.Errorf("unsupported api version %q", version)
	}
	return version, nil
}

func newRequestID() string {
	counter := requestCounter.Add(1)
	raw := make([]byte, 8)
	now := uint64(time.Now().UnixNano()) ^ counter
	for i := range raw {
		raw[len(raw)-1-i] = byte(now >> (i * 8))
	}
	return "req_" + hex.EncodeToString(raw)
}

func (r GenerateRequest) ResolvedWriteFiles() bool {
	if r.WriteFiles != nil {
		return *r.WriteFiles
	}
	return !r.DryRun
}

func (r GenerateRequest) ResolvedDryRun() bool {
	return !r.ResolvedWriteFiles()
}

func writeMode(writeFiles bool) string {
	if writeFiles {
		return "write_files"
	}
	return "dry_run"
}

func newGenerateResponse(req GenerateRequest, targetPath string) *GenerateResponse {
	writeFiles := req.ResolvedWriteFiles()
	return &GenerateResponse{
		APIVersion: APIVersion,
		RequestID:  req.RequestID,
		Success:    true,
		DryRun:     !writeFiles,
		WriteFiles: writeFiles,
		WriteMode:  writeMode(writeFiles),
		TargetPath: targetPath,
		Results:    []*models.GenerationResult{},
		Artifacts:  []Artifact{},
		Patches:    []PatchOperation{},
	}
}

func newAnalyzeResponse(req AnalyzeRequest, targetPath string) *AnalyzeResponse {
	return &AnalyzeResponse{
		APIVersion: APIVersion,
		RequestID:  req.RequestID,
		Success:    true,
		Path:       targetPath,
		ByLanguage: make(map[string]LangStats),
		Files:      []FileAnalysis{},
	}
}

func newValidateResponse(req ValidateRequest, targetPath string) *ValidateResponse {
	return &ValidateResponse{
		APIVersion: APIVersion,
		RequestID:  req.RequestID,
		Success:    true,
		TargetPath: targetPath,
	}
}

// NewGenerateFailureResponse builds a structured machine-readable failure envelope.
func NewGenerateFailureResponse(req GenerateRequest, err error, targetPath string) *GenerateResponse {
	resp := newGenerateResponse(req, targetPath)
	resp.Success = false
	resp.FailureCode = classifyFailure(err)
	resp.Error = errorMessage(err)
	return resp
}

// NewAnalyzeFailureResponse builds a structured machine-readable failure envelope.
func NewAnalyzeFailureResponse(req AnalyzeRequest, err error, targetPath string) *AnalyzeResponse {
	resp := newAnalyzeResponse(req, targetPath)
	resp.Success = false
	resp.FailureCode = classifyFailure(err)
	resp.Error = errorMessage(err)
	return resp
}

// NewValidateFailureResponse builds a structured machine-readable failure envelope.
func NewValidateFailureResponse(req ValidateRequest, err error, targetPath string) *ValidateResponse {
	resp := newValidateResponse(req, targetPath)
	resp.Success = false
	resp.FailureCode = classifyFailure(err)
	resp.Error = errorMessage(err)
	return resp
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func classifyFailure(err error) FailureCode {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return FailureCodeCancelled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return FailureCodeProviderTimeout
	}
	if errors.Is(err, llm.ErrNoAPIKey) {
		return FailureCodeMissingAPIKey
	}
	if errors.Is(err, llm.ErrRateLimited) {
		return FailureCodeProviderRateLimited
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "unsupported api version"),
		strings.Contains(lower, "either path or file is required"),
		strings.Contains(lower, "failed to resolve path"),
		strings.Contains(lower, "invalid generate request"):
		return FailureCodeInvalidRequest
	case strings.Contains(lower, "unsupported language"),
		strings.Contains(lower, "no adapter for language"):
		return FailureCodeUnsupportedLanguage
	case strings.Contains(lower, "no source files found"):
		return FailureCodeNoSourceFiles
	case strings.Contains(lower, "api key not configured"):
		return FailureCodeMissingAPIKey
	case strings.Contains(lower, "rate limited"),
		strings.Contains(lower, "status 429"):
		return FailureCodeProviderRateLimited
	case strings.Contains(lower, "timeout"),
		strings.Contains(lower, "status 504"),
		strings.Contains(lower, "status 503"),
		strings.Contains(lower, "status 502"),
		strings.Contains(lower, "status 500"):
		return FailureCodeProviderTimeout
	case strings.Contains(lower, "empty response from provider"),
		strings.Contains(lower, "malformed provider output"):
		return FailureCodeMalformedProviderOutput
	case strings.Contains(lower, "validation failed"):
		return FailureCodeValidationFailed
	case strings.Contains(lower, "failed to write test file"),
		strings.Contains(lower, "write test file"):
		return FailureCodeWriteFailed
	case strings.Contains(lower, "failed to scan path"):
		return FailureCodeScanFailed
	default:
		return FailureCodeInternalError
	}
}

func targetPathHint(pathValue string, fileValue string) string {
	target := pathValue
	if fileValue != "" {
		target = fileValue
	}
	if target == "" {
		return ""
	}
	absPath, err := filepath.Abs(target)
	if err != nil {
		return target
	}
	return absPath
}
