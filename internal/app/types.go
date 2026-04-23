package app

import (
	"github.com/princepal9120/testgen-cli/internal/llm"
	"github.com/princepal9120/testgen-cli/internal/validation"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

// GenerateRequest defines a machine-readable test generation request.
type GenerateRequest struct {
	APIVersion     string   `json:"api_version,omitempty"`
	RequestID      string   `json:"request_id,omitempty"`
	Path           string   `json:"path,omitempty"`
	File           string   `json:"file,omitempty"`
	Recursive      bool     `json:"recursive,omitempty"`
	IncludePattern string   `json:"include_pattern,omitempty"`
	ExcludePattern string   `json:"exclude_pattern,omitempty"`
	TestTypes      []string `json:"test_types,omitempty"`
	Framework      string   `json:"framework,omitempty"`
	OutputDir      string   `json:"output_dir,omitempty"`
	DryRun         bool     `json:"dry_run,omitempty"`
	WriteFiles     *bool    `json:"write_files,omitempty"`
	Validate       bool     `json:"validate,omitempty"`
	BatchSize      int      `json:"batch_size,omitempty"`
	Parallelism    int      `json:"parallelism,omitempty"`
	Provider       string   `json:"provider,omitempty"`
	ReportUsage    bool     `json:"report_usage,omitempty"`
	EmitPatch      bool     `json:"emit_patch,omitempty"`
}

// GenerateResponse contains the shared generation result returned to callers.
type GenerateResponse struct {
	APIVersion     string                     `json:"api_version"`
	RequestID      string                     `json:"request_id,omitempty"`
	Success        bool                       `json:"success"`
	FailureCode    FailureCode                `json:"failure_code,omitempty"`
	Error          string                     `json:"error,omitempty"`
	DryRun         bool                       `json:"dry_run"`
	WriteFiles     bool                       `json:"write_files"`
	WriteMode      string                     `json:"write_mode,omitempty"`
	TargetPath     string                     `json:"target_path"`
	SourceFiles    []*models.SourceFile       `json:"source_files,omitempty"`
	Results        []*models.GenerationResult `json:"results"`
	Artifacts      []Artifact                 `json:"artifacts,omitempty"`
	Patches        []PatchOperation           `json:"patches,omitempty"`
	Usage          *UsageReport               `json:"usage,omitempty"`
	SuccessCount   int                        `json:"success_count"`
	ErrorCount     int                        `json:"error_count"`
	TotalFunctions int                        `json:"total_functions"`
	Usage          *llm.UsageMetrics          `json:"usage,omitempty"`
}

// UsageReport is additive runtime accounting for a generation request.
type UsageReport struct {
	RequestCount     int     `json:"request_count"`
	TotalTokensIn    int     `json:"total_tokens_in"`
	TotalTokensOut   int     `json:"total_tokens_out"`
	CacheHits        int     `json:"cache_hits"`
	CacheMisses      int     `json:"cache_misses"`
	CachedTokens     int     `json:"cached_tokens"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// Artifact is a machine-readable generated artifact.
type Artifact struct {
	SourcePath       string      `json:"source_path"`
	Language         string      `json:"language"`
	TestPath         string      `json:"test_path,omitempty"`
	TestCode         string      `json:"test_code,omitempty"`
	FunctionsTested  []string    `json:"functions_tested,omitempty"`
	Generated        bool        `json:"generated"`
	FailureCode      FailureCode `json:"failure_code,omitempty"`
	Error            string      `json:"error,omitempty"`
	ValidationFailed bool        `json:"validation_failed,omitempty"`
}

// PatchOperation is a structured write operation suitable for agent wrappers.
type PatchOperation struct {
	Path    string `json:"path"`
	Action  string `json:"action"`
	Content string `json:"content"`
}

// AnalyzeRequest defines a machine-readable analyze request.
type AnalyzeRequest struct {
	APIVersion   string `json:"api_version,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	Path         string `json:"path,omitempty"`
	Recursive    bool   `json:"recursive,omitempty"`
	CostEstimate bool   `json:"cost_estimate,omitempty"`
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	BatchSize    int    `json:"batch_size,omitempty"`
	Detail       string `json:"detail,omitempty"`
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	BatchSize    int    `json:"batch_size,omitempty"`
}

// AnalyzeResponse contains analysis details for a codebase.
type AnalyzeResponse struct {
	APIVersion             string               `json:"api_version"`
	RequestID              string               `json:"request_id,omitempty"`
	Success                bool                 `json:"success"`
	FailureCode            FailureCode          `json:"failure_code,omitempty"`
	Error                  string               `json:"error,omitempty"`
	Path                   string               `json:"path"`
	TotalFiles             int                  `json:"total_files"`
	TotalFunctions         int                  `json:"total_functions"`
	TotalLines             int                  `json:"total_lines"`
	ByLanguage             map[string]LangStats `json:"by_language"`
	ExactFunctionFiles     int                  `json:"exact_function_files,omitempty"`
	HeuristicFunctionFiles int                  `json:"heuristic_function_files,omitempty"`
	Provider               string               `json:"provider,omitempty"`
	Model                  string               `json:"model,omitempty"`
	EstimatedRequests      int                  `json:"estimated_requests,omitempty"`
	EstimatedBatchCount    int                  `json:"estimated_batch_count,omitempty"`
	EstimatedChunkCount    int                  `json:"estimated_chunk_count,omitempty"`
	EstimatedInputTokens   int                  `json:"estimated_input_tokens,omitempty"`
	EstimatedOutputTokens  int                  `json:"estimated_output_tokens,omitempty"`
	EstimatedTokens        int                  `json:"estimated_tokens,omitempty"`
	EstimatedCost          float64              `json:"estimated_cost_usd,omitempty"`
	InputCostPerMTokens    float64              `json:"input_cost_per_million_usd,omitempty"`
	OutputCostPerMTokens   float64              `json:"output_cost_per_million_usd,omitempty"`
	CostEstimateOffline    bool                 `json:"cost_estimate_offline,omitempty"`
	Warnings               []string             `json:"warnings,omitempty"`
	Usage                  *llm.UsageMetrics    `json:"usage,omitempty"`
	Files                  []FileAnalysis       `json:"files,omitempty"`
}

// LangStats captures aggregate stats per language.
type LangStats struct {
	Files     int `json:"files"`
	Lines     int `json:"lines"`
	Functions int `json:"functions"`
}

// FileAnalysis captures per-file analysis output.
type FileAnalysis struct {
	Path              string `json:"path"`
	Language          string `json:"language"`
	Lines             int    `json:"lines"`
	Functions         int    `json:"functions"`
	FunctionCountMode string `json:"function_count_mode,omitempty"`
	EstimatedRequests int    `json:"estimated_requests,omitempty"`
	EstimatedBatches  int    `json:"estimated_batch_count,omitempty"`
	InputTokens       int    `json:"estimated_input_tokens,omitempty"`
	OutputTokens      int    `json:"estimated_output_tokens,omitempty"`
	Tokens            int    `json:"estimated_tokens,omitempty"`
	EstimatedCost     float64 `json:"estimated_cost_usd,omitempty"`
}

// ValidateRequest defines a machine-readable validate request.
type ValidateRequest struct {
	APIVersion    string  `json:"api_version,omitempty"`
	RequestID     string  `json:"request_id,omitempty"`
	Path          string  `json:"path,omitempty"`
	Recursive     bool    `json:"recursive,omitempty"`
	MinCoverage   float64 `json:"min_coverage,omitempty"`
	FailOnMissing bool    `json:"fail_on_missing,omitempty"`
	ReportGaps    bool    `json:"report_gaps,omitempty"`
}

// ValidateResponse contains validation output plus scan metadata.
type ValidateResponse struct {
	APIVersion  string               `json:"api_version"`
	RequestID   string               `json:"request_id,omitempty"`
	Success     bool                 `json:"success"`
	FailureCode FailureCode          `json:"failure_code,omitempty"`
	Error       string               `json:"error,omitempty"`
	TargetPath  string               `json:"target_path"`
	SourceFiles []*models.SourceFile `json:"source_files,omitempty"`
	Result      *validation.Result   `json:"result"`
}
