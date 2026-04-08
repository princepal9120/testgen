package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/generator"
	"github.com/princepal9120/testgen-cli/internal/scanner"
	"github.com/princepal9120/testgen-cli/internal/validation"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

// Service centralizes machine-readable orchestration for CLI, TUI, and future wrappers.
type Service struct {
	registry     *adapters.Registry
	newEngine    func(generator.EngineConfig) (*generator.Engine, error)
	newScanner   func(scanner.Options) *scanner.Scanner
	newValidator func(validation.Config) *validation.Validator
}

// NewService creates a Service with production dependencies.
func NewService() *Service {
	return &Service{
		registry:     adapters.DefaultRegistry(),
		newEngine:    generator.NewEngine,
		newScanner:   scanner.New,
		newValidator: validation.NewValidator,
	}
}

// Generate executes test generation via the shared application layer.
func (s *Service) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	targetPath, err := resolveTargetPath(req.Path, req.File)
	if err != nil {
		return nil, err
	}

	scannerOpts := scanner.Options{
		Recursive:      req.Recursive,
		IncludePattern: req.IncludePattern,
		ExcludePattern: req.ExcludePattern,
	}

	sourceFiles, err := s.newScanner(scannerOpts).Scan(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan path: %w", err)
	}
	if len(sourceFiles) == 0 {
		return nil, fmt.Errorf("no source files found")
	}

	engine, err := s.newEngine(generator.EngineConfig{
		DryRun:      req.DryRun,
		Validate:    req.Validate,
		OutputDir:   req.OutputDir,
		TestTypes:   req.TestTypes,
		Framework:   req.Framework,
		BatchSize:   req.BatchSize,
		Parallelism: req.Parallelism,
		Provider:    req.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize generator: %w", err)
	}

	results := s.generateResults(ctx, sourceFiles, engine, req.Parallelism)

	resp := &GenerateResponse{
		TargetPath:  targetPath,
		SourceFiles: sourceFiles,
		Results:     results,
		Artifacts:   make([]Artifact, 0, len(results)),
	}
	for _, result := range results {
		if result == nil {
			continue
		}
		if !req.DryRun {
			adapter := s.registry.GetAdapter(result.SourceFile.Language)
			if adapter == nil {
				result.Error = fmt.Errorf("no adapter for language: %s", result.SourceFile.Language)
				result.ErrorMessage = result.Error.Error()
			} else if err := engine.MaterializeResult(result, adapter); err != nil {
				result.Error = err
				result.ErrorMessage = err.Error()
			}
		}

		artifact := artifactFromResult(result)
		if artifact.Generated || artifact.Error != "" {
			resp.Artifacts = append(resp.Artifacts, artifact)
		}

		if req.DryRun || req.EmitPatch {
			if patch := patchFromResult(result); patch != nil {
				resp.Patches = append(resp.Patches, *patch)
			}
		}

		if result.Error != nil {
			resp.ErrorCount++
		} else {
			resp.SuccessCount++
		}
		resp.TotalFunctions += len(result.FunctionsTested)
	}

	return resp, nil
}

// Analyze executes codebase analysis via the shared application layer.
func (s *Service) Analyze(_ context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	targetPath, err := resolveTargetPath(req.Path, "")
	if err != nil {
		return nil, err
	}

	sourceFiles, err := s.newScanner(scanner.Options{Recursive: req.Recursive}).Scan(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan path: %w", err)
	}

	result := analyzeFiles(sourceFiles, targetPath)
	if req.CostEstimate {
		estimateCosts(result)
	}
	if req.Detail == "summary" {
		result.Files = nil
	}

	return result, nil
}

// Validate executes validation via the shared application layer.
func (s *Service) Validate(_ context.Context, req ValidateRequest) (*ValidateResponse, error) {
	targetPath, err := resolveTargetPath(req.Path, "")
	if err != nil {
		return nil, err
	}

	sourceFiles, err := s.newScanner(scanner.Options{Recursive: req.Recursive}).Scan(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan path: %w", err)
	}

	validator := s.newValidator(validation.Config{
		MinCoverage:   req.MinCoverage,
		FailOnMissing: req.FailOnMissing,
		ReportGaps:    req.ReportGaps,
	})

	result, err := validator.Validate(targetPath, sourceFiles)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &ValidateResponse{
		TargetPath:  targetPath,
		SourceFiles: sourceFiles,
		Result:      result,
	}, nil
}

func resolveTargetPath(path string, file string) (string, error) {
	targetPath := path
	if file != "" {
		targetPath = file
	}
	if targetPath == "" {
		return "", fmt.Errorf("either path or file is required")
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	return absPath, nil
}

func (s *Service) generateResults(ctx context.Context, files []*models.SourceFile, engine *generator.Engine, parallelism int) []*models.GenerationResult {
	if parallelism > 1 {
		pool := generator.NewWorkerPool(engine, parallelism)
		results := pool.ProcessFiles(ctx, files)
		sortGenerationResults(results)
		return results
	}

	results := make([]*models.GenerationResult, 0, len(files))
	for _, file := range files {
		select {
		case <-ctx.Done():
			return append(results, &models.GenerationResult{
				SourceFile:   file,
				Error:        ctx.Err(),
				ErrorMessage: ctx.Err().Error(),
			})
		default:
		}

		adapter := s.registry.GetAdapter(file.Language)
		if adapter == nil {
			results = append(results, &models.GenerationResult{
				SourceFile:   file,
				Error:        fmt.Errorf("no adapter for language: %s", file.Language),
				ErrorMessage: "no adapter for language: " + file.Language,
			})
			continue
		}

		result, err := engine.GenerateArtifact(file, adapter)
		if err != nil {
			results = append(results, &models.GenerationResult{
				SourceFile:   file,
				Error:        err,
				ErrorMessage: err.Error(),
			})
			continue
		}

		results = append(results, result)
	}

	sortGenerationResults(results)
	return results
}

func sortGenerationResults(results []*models.GenerationResult) {
	sort.Slice(results, func(i, j int) bool {
		left := ""
		right := ""
		if results[i] != nil && results[i].SourceFile != nil {
			left = results[i].SourceFile.Path
		}
		if results[j] != nil && results[j].SourceFile != nil {
			right = results[j].SourceFile.Path
		}
		return left < right
	})
}

func artifactFromResult(result *models.GenerationResult) Artifact {
	artifact := Artifact{}
	if result == nil || result.SourceFile == nil {
		return artifact
	}

	artifact.SourcePath = result.SourceFile.Path
	artifact.Language = result.SourceFile.Language
	artifact.TestPath = result.TestPath
	artifact.TestCode = result.TestCode
	artifact.FunctionsTested = result.FunctionsTested
	artifact.Generated = result.TestCode != ""
	if result.Error != nil {
		artifact.Error = result.Error.Error()
		artifact.ValidationFailed = strings.Contains(result.Error.Error(), "validation failed")
	}
	if result.ErrorMessage != "" && artifact.Error == "" {
		artifact.Error = result.ErrorMessage
	}

	return artifact
}

func patchFromResult(result *models.GenerationResult) *PatchOperation {
	if result == nil || result.TestPath == "" || result.TestCode == "" {
		return nil
	}

	action := "create_or_replace"
	if _, err := os.Stat(result.TestPath); err == nil {
		action = "replace"
	}

	return &PatchOperation{
		Path:    result.TestPath,
		Action:  action,
		Content: result.TestCode,
	}
}

func analyzeFiles(files []*scanner.SourceFile, basePath string) *AnalyzeResponse {
	result := &AnalyzeResponse{
		Path:       basePath,
		ByLanguage: make(map[string]LangStats),
		Files:      make([]FileAnalysis, 0, len(files)),
	}

	for _, f := range files {
		content, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}

		lines := len(strings.Split(string(content), "\n"))
		estimatedFunctions := max(1, lines/20)

		result.TotalFiles++
		result.TotalLines += lines
		result.TotalFunctions += estimatedFunctions

		stats := result.ByLanguage[f.Language]
		stats.Files++
		stats.Lines += lines
		stats.Functions += estimatedFunctions
		result.ByLanguage[f.Language] = stats

		relPath, _ := filepath.Rel(basePath, f.Path)
		result.Files = append(result.Files, FileAnalysis{
			Path:      relPath,
			Language:  f.Language,
			Lines:     lines,
			Functions: estimatedFunctions,
		})
	}

	return result
}

func estimateCosts(result *AnalyzeResponse) {
	tokensPerFunction := 150
	outputPerFunction := 200
	batchSize := 5
	systemPromptTokens := 500

	totalInputTokens := (result.TotalFunctions * tokensPerFunction) +
		((result.TotalFunctions / batchSize) * systemPromptTokens)
	totalOutputTokens := result.TotalFunctions * outputPerFunction

	result.EstimatedTokens = totalInputTokens + totalOutputTokens
	inputCost := float64(totalInputTokens) * 3.00 / 1_000_000
	outputCost := float64(totalOutputTokens) * 15.00 / 1_000_000
	result.EstimatedCost = inputCost + outputCost
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
