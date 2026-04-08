package app

import (
	"github.com/princepal9120/testgen-cli/internal/validation"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

// GenerateRequest defines a machine-readable test generation request.
type GenerateRequest struct {
	Path           string
	File           string
	Recursive      bool
	IncludePattern string
	ExcludePattern string
	TestTypes      []string
	Framework      string
	OutputDir      string
	DryRun         bool
	Validate       bool
	BatchSize      int
	Parallelism    int
	Provider       string
	EmitPatch      bool
}

// GenerateResponse contains the shared generation result returned to callers.
type GenerateResponse struct {
	TargetPath     string                     `json:"target_path"`
	SourceFiles    []*models.SourceFile       `json:"source_files,omitempty"`
	Results        []*models.GenerationResult `json:"results"`
	Artifacts      []Artifact                 `json:"artifacts,omitempty"`
	Patches        []PatchOperation           `json:"patches,omitempty"`
	SuccessCount   int                        `json:"success_count"`
	ErrorCount     int                        `json:"error_count"`
	TotalFunctions int                        `json:"total_functions"`
}

// Artifact is a machine-readable generated artifact.
type Artifact struct {
	SourcePath       string   `json:"source_path"`
	Language         string   `json:"language"`
	TestPath         string   `json:"test_path,omitempty"`
	TestCode         string   `json:"test_code,omitempty"`
	FunctionsTested  []string `json:"functions_tested,omitempty"`
	Generated        bool     `json:"generated"`
	Error            string   `json:"error,omitempty"`
	ValidationFailed bool     `json:"validation_failed,omitempty"`
}

// PatchOperation is a structured write operation suitable for agent wrappers.
type PatchOperation struct {
	Path    string `json:"path"`
	Action  string `json:"action"`
	Content string `json:"content"`
}

// AnalyzeRequest defines a machine-readable analyze request.
type AnalyzeRequest struct {
	Path         string
	Recursive    bool
	CostEstimate bool
	Detail       string
}

// AnalyzeResponse contains analysis details for a codebase.
type AnalyzeResponse struct {
	Path            string               `json:"path"`
	TotalFiles      int                  `json:"total_files"`
	TotalFunctions  int                  `json:"total_functions"`
	TotalLines      int                  `json:"total_lines"`
	ByLanguage      map[string]LangStats `json:"by_language"`
	EstimatedTokens int                  `json:"estimated_tokens,omitempty"`
	EstimatedCost   float64              `json:"estimated_cost_usd,omitempty"`
	Files           []FileAnalysis       `json:"files,omitempty"`
}

// LangStats captures aggregate stats per language.
type LangStats struct {
	Files     int `json:"files"`
	Lines     int `json:"lines"`
	Functions int `json:"functions"`
}

// FileAnalysis captures per-file analysis output.
type FileAnalysis struct {
	Path      string `json:"path"`
	Language  string `json:"language"`
	Lines     int    `json:"lines"`
	Functions int    `json:"functions"`
	Tokens    int    `json:"estimated_tokens,omitempty"`
}

// ValidateRequest defines a machine-readable validate request.
type ValidateRequest struct {
	Path          string
	Recursive     bool
	MinCoverage   float64
	FailOnMissing bool
	ReportGaps    bool
}

// ValidateResponse contains validation output plus scan metadata.
type ValidateResponse struct {
	TargetPath  string               `json:"target_path"`
	SourceFiles []*models.SourceFile `json:"source_files,omitempty"`
	Result      *validation.Result   `json:"result"`
}
