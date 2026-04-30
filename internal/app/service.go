package app

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/generator"
	"github.com/princepal9120/testgen-cli/internal/llm"
	"github.com/princepal9120/testgen-cli/internal/metrics"
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
	if err := normalizeGenerateRequest(&req); err != nil {
		return nil, err
	}

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
		if req.File != "" && scanner.DetectLanguage(targetPath) == "" {
			return nil, fmt.Errorf("unsupported language for file: %s", filepath.Ext(targetPath))
		}
		return nil, fmt.Errorf("no source files found")
	}

	engine, err := s.newEngine(generator.EngineConfig{
		DryRun:      req.ResolvedDryRun(),
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

	resp := newGenerateResponse(req, targetPath)
	resp.SourceFiles = sourceFiles
	resp.Results = results
	resp.Artifacts = make([]Artifact, 0, len(results))
	resolvedDryRun := req.ResolvedDryRun()
	for _, result := range results {
		if result == nil {
			continue
		}
		if !resolvedDryRun {
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

		if resolvedDryRun || req.EmitPatch {
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
	resp.Usage = engine.GetUsage()
	resp.Success = resp.ErrorCount == 0
	if !resp.Success {
		for _, result := range results {
			if result == nil || result.Error == nil {
				continue
			}
			resp.FailureCode = classifyFailure(result.Error)
			resp.Error = result.Error.Error()
			break
		}
	}
	persistGenerateMetrics(targetPath, req.ReportUsage, resp)

	return resp, nil
}

// Analyze executes codebase analysis via the shared application layer.
func (s *Service) Analyze(_ context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	if err := normalizeAnalyzeRequest(&req); err != nil {
		return nil, err
	}

	targetPath, err := resolveTargetPath(req.Path, "")
	if err != nil {
		return nil, err
	}

	sourceFiles, err := s.newScanner(scanner.Options{Recursive: req.Recursive}).Scan(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan path: %w", err)
	}

	result := newAnalyzeResponse(req, targetPath)
	analyzeFiles(result, sourceFiles, targetPath, s.registry)
	if req.CostEstimate {
		estimateCosts(result, req)
	}
	if req.Detail == "summary" {
		result.Files = nil
	}
	persistAnalyzeMetrics(targetPath, result)

	return result, nil
}

// Validate executes validation via the shared application layer.
func (s *Service) Validate(_ context.Context, req ValidateRequest) (*ValidateResponse, error) {
	if err := normalizeValidateRequest(&req); err != nil {
		return nil, err
	}

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

	resp := newValidateResponse(req, targetPath)
	resp.SourceFiles = sourceFiles
	resp.Result = result
	persistValidationMetrics(targetPath, len(sourceFiles), result)
	if result != nil && len(result.Errors) > 0 {
		resp.Success = false
		resp.FailureCode = FailureCodeValidationFailed
		resp.Error = strings.Join(result.Errors, "; ")
	}

	return resp, nil
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
		artifact.FailureCode = classifyFailure(result.Error)
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

func buildGenerateUsageSummary(engine *generator.Engine) *llm.UsageMetrics {
	if engine == nil {
		return nil
	}

	usage := engine.GetUsage()
	_, hits, misses, _ := engine.GetCacheStats()
	if usage == nil {
		usage = &llm.UsageMetrics{}
	}

	provider := usage.Provider
	if provider == "" {
		provider = engine.GetProviderName()
	}
	model := usage.Model
	if model == "" {
		model = llm.GetDefaultModel(provider)
	}

	return &llm.UsageMetrics{
		Provider:         provider,
		Model:            model,
		TotalRequests:    usage.TotalRequests,
		BatchCount:       usage.BatchCount,
		ChunkCount:       usage.ChunkCount,
		TotalTokensIn:    usage.TotalTokensIn,
		TotalTokensOut:   usage.TotalTokensOut,
		CachedTokens:     usage.CachedTokens,
		EstimatedCostUSD: usage.EstimatedCostUSD,
		CacheHits:        hits,
		CacheMisses:      misses,
		Estimated:        usage.Estimated,
	}
}

func analyzeFiles(result *AnalyzeResponse, files []*scanner.SourceFile, basePath string, registry *adapters.Registry) {
	for _, f := range files {
		content, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}

		lines := len(strings.Split(string(content), "\n"))
		functionCount, countMode, warning := countFunctions(f, string(content), registry)

		result.TotalFiles++
		result.TotalLines += lines
		result.TotalFunctions += functionCount
		if countMode == "exact" {
			result.ExactFunctionFiles++
		} else {
			result.HeuristicFunctionFiles++
			if warning != "" {
				result.Warnings = append(result.Warnings, warning)
			}
		}

		stats := result.ByLanguage[f.Language]
		stats.Files++
		stats.Lines += lines
		stats.Functions += functionCount
		result.ByLanguage[f.Language] = stats

		relPath, _ := filepath.Rel(basePath, f.Path)
		result.Files = append(result.Files, FileAnalysis{
			Path:              relPath,
			Language:          f.Language,
			Lines:             lines,
			Functions:         functionCount,
			FunctionCountMode: countMode,
		})
	}

	sort.Strings(result.Warnings)
}

func estimateCosts(result *AnalyzeResponse, req AnalyzeRequest) {
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 5
	}

	provider := llm.ResolveProvider(req.Provider)
	if provider == "" {
		provider = "anthropic"
	}
	model := llm.ResolveModel(provider, req.Model)
	usage := &llm.UsageMetrics{
		Provider:  provider,
		Model:     model,
		Estimated: true,
	}

	for idx := range result.Files {
		file := &result.Files[idx]
		if file.Functions == 0 {
			continue
		}
		inputTokens, outputTokens, requests := estimateFileTokens(*file, batchSize)
		file.Tokens = inputTokens + outputTokens
		file.EstimatedCost = llm.EstimateCost(provider, model, inputTokens, outputTokens)

		usage.TotalRequests += requests
		usage.BatchCount += requests
		usage.ChunkCount += requests
		usage.TotalTokensIn += inputTokens
		usage.TotalTokensOut += outputTokens
		usage.EstimatedCostUSD += file.EstimatedCost
	}

	rateEstimate := llm.EstimateOfflineUsage(provider, model, 0, batchSize)
	result.Provider = provider
	result.Model = model
	result.EstimatedRequests = usage.TotalRequests
	result.EstimatedBatchCount = usage.BatchCount
	result.EstimatedChunkCount = usage.ChunkCount
	result.EstimatedInputTokens = usage.TotalTokensIn
	result.EstimatedOutputTokens = usage.TotalTokensOut
	result.EstimatedTokens = usage.TotalTokens()
	result.EstimatedCost = usage.EstimatedCostUSD
	result.InputCostPerMTokens = rateEstimate.InputCostPerMillionUSD
	result.OutputCostPerMTokens = rateEstimate.OutputCostPerMillionUSD
	result.CostEstimateOffline = true
	result.Usage = usage
}

func estimateFunctionTokens(functionCount int) (inputTokens int, outputTokens int) {
	if functionCount <= 0 {
		return 0, 0
	}

	const (
		tokensPerFunction  = 150
		outputPerFunction  = 200
		batchSize          = 5
		systemPromptTokens = 500
	)

	inputTokens = (functionCount * tokensPerFunction) +
		(((functionCount-1)/batchSize)+1)*systemPromptTokens
	outputTokens = functionCount * outputPerFunction
	return inputTokens, outputTokens
}

func usageReportFromEngine(engine *generator.Engine) *llm.UsageMetrics {
	if engine == nil {
		return &llm.UsageMetrics{}
	}
	return engine.GetUsage()
}

func countFunctions(file *scanner.SourceFile, content string, registry *adapters.Registry) (int, string, string) {
	if file == nil {
		return 0, "heuristic", ""
	}

	if registry != nil {
		if adapter := registry.GetAdapter(file.Language); adapter != nil {
			ast, err := adapter.ParseFile(content)
			if err == nil && ast != nil {
				definitions, err := adapter.ExtractDefinitions(ast)
				if err == nil {
					return len(definitions), "exact", ""
				}
			}
		}
	}

	return heuristicFunctionCount(file.Language, content), "heuristic", fmt.Sprintf("%s: function count fell back to heuristic estimation", file.Path)
}

func heuristicFunctionCount(language string, content string) int {
	patterns := map[string]string{
		"go":         `(?m)^func\s+(?:\([^)]*\)\s*)?[A-Za-z_]\w*\s*\(`,
		"python":     `(?m)^\s*(?:async\s+)?def\s+[A-Za-z_]\w*\s*\(`,
		"javascript": `(?m)(?:^|\s)(?:async\s+)?function\s+[A-Za-z_]\w*\s*\(|(?:const|let|var)\s+[A-Za-z_]\w*\s*=\s*(?:async\s*)?\([^)]*\)\s*=>`,
		"typescript": `(?m)(?:^|\s)(?:async\s+)?function\s+[A-Za-z_]\w*\s*\(|(?:const|let|var)\s+[A-Za-z_]\w*\s*=\s*(?:async\s*)?\([^)]*\)\s*=>`,
		"java":       `(?m)^\s*(?:public|private|protected|static|final|synchronized|abstract|\s)+[\w<>\[\], ?]+\s+[A-Za-z_]\w*\s*\(`,
		"rust":       `(?m)^\s*(?:pub\s+)?fn\s+[A-Za-z_]\w*\s*\(`,
	}

	pattern := patterns[scanner.NormalizeLanguage(language)]
	if pattern != "" {
		return len(regexp.MustCompile(pattern).FindAllStringIndex(content, -1))
	}

	lines := len(strings.Split(content, "\n"))
	if strings.TrimSpace(content) == "" {
		return 0
	}
	return max(1, lines/40)
}

func persistAnalyzeMetrics(targetPath string, result *AnalyzeResponse) {
	if result == nil {
		return
	}

	collector := metrics.NewCollector()
	collector.SetContext("analyze", targetPath, false)
	collector.SetAnalyzeSummary(result.TotalFiles, result.ExactFunctionFiles, result.HeuristicFunctionFiles)
	if result.Usage != nil {
		collector.ApplyUsage(result.Usage)
	} else {
		if result.EstimatedTokens > 0 {
			collector.RecordTokens(result.EstimatedTokens, 0, false)
		}
		if result.EstimatedCost > 0 {
			collector.RecordCost(result.EstimatedCost)
		}
	}
	if err := collector.Save(); err != nil {
		slog.Warn("failed to persist analyze metrics", slog.String("path", targetPath), slog.String("error", err.Error()))
	}
}

func persistGenerateMetrics(targetPath string, reportUsage bool, result *GenerateResponse) {
	if result == nil || (!reportUsage && result.Usage == nil) {
		return
	}

	collector := metrics.NewCollector()
	collector.SetContext("generate", targetPath, false)
	collector.SetGenerateSummary(len(result.SourceFiles), result.SuccessCount, result.ErrorCount)
	if result.Usage != nil {
		collector.ApplyUsage(result.Usage)
	}
	if err := collector.Save(); err != nil {
		slog.Warn("failed to persist generate metrics", slog.String("path", targetPath), slog.String("error", err.Error()))
	}
}

func persistValidationMetrics(targetPath string, totalFiles int, result *validation.Result) {
	if result == nil {
		return
	}

	collector := metrics.NewCollector()
	collector.SetContext("validate", targetPath, false)
	collector.SetValidationSummary(totalFiles, result.CoveragePercent, result.TestsPassed, result.TestsFailed, len(result.FilesMissingTests), len(result.Errors))
	if err := collector.Save(); err != nil {
		slog.Warn("failed to persist validation metrics", slog.String("path", targetPath), slog.String("error", err.Error()))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func estimateFileTokens(file FileAnalysis, batchSize int) (inputTokens int, outputTokens int, requests int) {
	if file.Functions <= 0 {
		return 0, 0, 0
	}

	requests = int(math.Ceil(float64(file.Functions) / float64(max(batchSize, 1))))
	tokensPerFunction := 120 + max(file.Lines/max(file.Functions, 1), 1)*4
	outputPerFunction := 200
	systemPromptTokens := 250
	inputTokens = (file.Functions * tokensPerFunction) + (requests * systemPromptTokens)
	outputTokens = file.Functions * outputPerFunction
	return inputTokens, outputTokens, requests
}
