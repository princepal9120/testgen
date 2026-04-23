/*
Package generator implements the core test generation engine.

This package orchestrates the test generation process by coordinating
language adapters, LLM providers, and output formatting.
*/
package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/llm"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

// EngineConfig contains configuration for the generation engine
type EngineConfig struct {
	DryRun      bool
	Validate    bool
	OutputDir   string
	TestTypes   []string
	Framework   string
	BatchSize   int
	Parallelism int
	Provider    string // "anthropic" or "openai"
}

// Engine orchestrates test generation
type Engine struct {
	config   EngineConfig
	provider llm.Provider
	cache    *llm.Cache
	logger   *slog.Logger
	usageMu  sync.Mutex
	usage    llm.UsageMetrics
}

// NewEngine creates a new generation engine
func NewEngine(config EngineConfig) (*Engine, error) {
	logger := slog.Default()

	// Initialize LLM provider
	var provider llm.Provider
	switch strings.ToLower(config.Provider) {
	case "openai":
		provider = llm.NewOpenAIProvider()
	case "gemini":
		provider = llm.NewGeminiProvider()
	case "groq":
		provider = llm.NewGroqProvider()
	default:
		// Default to Anthropic
		provider = llm.NewAnthropicProvider()
	}

	// Configure provider
	if err := provider.Configure(llm.ProviderConfig{}); err != nil {
		// Not configured, will fail on actual generation
		logger.Warn("LLM provider not configured", slog.String("error", err.Error()))
	}
	provider = llm.NewReliableProvider(provider)

	return &Engine{
		config:   config,
		provider: provider,
		cache:    llm.NewCache(10000),
		logger:   logger,
		usage: llm.UsageMetrics{
			Provider: llm.ResolveProvider(config.Provider),
			Model:    llm.ResolveModel(config.Provider, ""),
		},
	}, nil
}

// GenerateArtifact generates test artifacts for a source file without writing them.
func (e *Engine) GenerateArtifact(sourceFile *models.SourceFile, adapter adapters.LanguageAdapter) (*models.GenerationResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	result := &models.GenerationResult{
		SourceFile: sourceFile,
	}

	// Read source file content
	content, err := os.ReadFile(sourceFile.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	// Parse file
	ast, err := adapter.ParseFile(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Extract definitions
	definitions, err := adapter.ExtractDefinitions(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to extract definitions: %w", err)
	}

	if len(definitions) == 0 {
		e.logger.Info("no functions found in file", slog.String("path", sourceFile.Path))
		return result, nil
	}

	e.logger.Debug("extracted definitions",
		slog.String("path", sourceFile.Path),
		slog.Int("count", len(definitions)),
	)

	// Generate tests for each definition
	tasks := buildGenerationTasks(e.provider, e.config, adapter, definitions, ast.Package)
	generated := make(map[string]string, len(tasks))
	var allTests strings.Builder
	functionsTested := make([]string, 0)
	var firstGenerationErr error

	pending := make([]generationTask, 0, len(tasks))
	for _, task := range tasks {
		if cached, hit := e.cache.Get(task.cacheKey); hit {
			e.logger.Debug("cache hit", slog.String("function", task.def.Name), slog.String("test_type", task.testType))
			e.recordCachedUsage(cached)
			generated[task.cacheKey] = cached.Content
			continue
		}
		pending = append(pending, task)
	}

	for _, chunk := range planGenerationChunks(e.provider, e.config, adapter, ast.Package, pending) {
		chunkCodes, err := e.generateChunk(ctx, chunk, adapter)
		if err != nil {
			if firstGenerationErr == nil {
				firstGenerationErr = err
			}
			e.logger.Warn("failed to generate chunk",
				slog.Int("definitions", len(chunk.tasks)),
				slog.String("error", err.Error()),
			)
			for _, task := range chunk.tasks {
				testCode, singleErr := e.generateSingleTask(ctx, task, adapter)
				if singleErr != nil {
					if firstGenerationErr == nil {
						firstGenerationErr = singleErr
					}
					e.logger.Warn("failed to generate test",
						slog.String("function", task.def.Name),
						slog.String("test_type", task.testType),
						slog.String("error", singleErr.Error()),
					)
					continue
				}
				generated[task.cacheKey] = testCode
			}
			continue
		}
		for key, code := range chunkCodes {
			generated[key] = code
		}
	}

	for _, task := range tasks {
		testCode := strings.TrimSpace(generated[task.cacheKey])
		if testCode == "" {
			continue
		}
		allTests.WriteString(testCode)
		allTests.WriteString("\n\n")
		functionsTested = append(functionsTested, task.def.Name)
	}

	if allTests.Len() == 0 {
		if firstGenerationErr != nil {
			result.Error = firstGenerationErr
			result.ErrorMessage = firstGenerationErr.Error()
		}
		return result, nil
	}

	// Post-process: add imports, format
	finalCode := e.postProcess(allTests.String(), adapter, sourceFile.Language, ast)

	// Format code
	formattedCode, err := adapter.FormatTestCode(finalCode)
	if err != nil {
		e.logger.Warn("failed to format test code", slog.String("error", err.Error()))
		formattedCode = finalCode
	}

	result.TestCode = formattedCode
	result.FunctionsTested = functionsTested
	result.TestCount = len(functionsTested)

	// Determine test file path
	testPath := adapter.GenerateTestPath(sourceFile.Path, e.config.OutputDir)
	result.TestPath = testPath

	return result, nil
}

// MaterializeResult writes and validates a generated result according to engine config.
func (e *Engine) MaterializeResult(result *models.GenerationResult, adapter adapters.LanguageAdapter) error {
	if result == nil || result.TestPath == "" || result.TestCode == "" {
		return nil
	}

	if err := e.writeTestFile(result.TestPath, result.TestCode); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}
	e.logger.Info("wrote test file", slog.String("path", result.TestPath))

	if e.config.Validate {
		if err := adapter.ValidateTests(result.TestCode, result.TestPath); err != nil {
			result.Error = fmt.Errorf("validation failed: %w", err)
			result.ErrorMessage = result.Error.Error()
			e.logger.Warn("test validation failed", slog.String("error", err.Error()))
		}
	}

	return nil
}

// Generate generates tests for a source file and materializes them when configured.
func (e *Engine) Generate(sourceFile *models.SourceFile, adapter adapters.LanguageAdapter) (*models.GenerationResult, error) {
	result, err := e.GenerateArtifact(sourceFile, adapter)
	if err != nil {
		return nil, err
	}

	if !e.config.DryRun {
		if err := e.MaterializeResult(result, adapter); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (e *Engine) generateSingleTask(ctx context.Context, task generationTask, adapter adapters.LanguageAdapter) (string, error) {
	resp, err := e.provider.Complete(ctx, llm.CompletionRequest{
		Prompt:      task.prompt,
		SystemRole:  defaultSystemRole(adapter),
		Temperature: 0.3,
		MaxTokens:   2000,
	})
	if err != nil {
		return "", fmt.Errorf("LLM completion failed: %w", err)
	}

	code := extractCodeFromResponse(resp.Content, adapter.GetLanguage())
	if strings.TrimSpace(code) == "" {
		return "", fmt.Errorf("empty response from provider")
	}

	resp.Content = code
	e.cache.Set(task.cacheKey, resp)
	e.recordLiveUsage(resp, 1)
	return code, nil
}

func (e *Engine) generateChunk(ctx context.Context, chunk generationChunk, adapter adapters.LanguageAdapter) (map[string]string, error) {
	if len(chunk.tasks) == 0 {
		return nil, nil
	}
	if len(chunk.tasks) == 1 {
		code, err := e.generateSingleTask(ctx, chunk.tasks[0], adapter)
		if err != nil {
			return nil, err
		}
		return map[string]string{chunk.tasks[0].cacheKey: code}, nil
	}

	resp, err := e.provider.Complete(ctx, llm.CompletionRequest{
		Prompt:      chunk.prompt,
		SystemRole:  chunk.systemRole,
		Temperature: 0.3,
		MaxTokens:   4000,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM batch completion failed: %w", err)
	}

	parsed, err := parseChunkResponse(resp.Content, adapter.GetLanguage(), chunk.tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse batched response: %w", err)
	}

	results := make(map[string]string, len(chunk.tasks))
	totalEstimated := 0
	for _, task := range chunk.tasks {
		totalEstimated += max(task.estimatedTokens, 1)
	}
	if totalEstimated == 0 {
		totalEstimated = len(chunk.tasks)
	}

	remainingIn := resp.TokensInput
	remainingOut := resp.TokensOutput
	remainingCost := resp.EstimatedCostUSD

	for idx, task := range chunk.tasks {
		code := strings.TrimSpace(parsed[task.id])
		if code == "" {
			return nil, fmt.Errorf("empty code for task id %s", task.id)
		}

		allocatedIn, allocatedOut, allocatedCost := splitChunkMetrics(resp, remainingIn, remainingOut, remainingCost, totalEstimated, task.estimatedTokens, idx == len(chunk.tasks)-1)
		remainingIn -= allocatedIn
		remainingOut -= allocatedOut
		remainingCost -= allocatedCost
		totalEstimated -= max(task.estimatedTokens, 1)

		cachedResp := &llm.CompletionResponse{
			Content:          code,
			TokensInput:      allocatedIn,
			TokensOutput:     allocatedOut,
			Provider:         firstNonEmpty(resp.Provider, e.provider.Name()),
			Model:            resp.Model,
			FinishReason:     resp.FinishReason,
			EstimatedCostUSD: allocatedCost,
		}
		e.cache.Set(task.cacheKey, cachedResp)
		results[task.cacheKey] = code
	}

	e.recordLiveUsage(resp, len(chunk.tasks))
	return results, nil
}

func splitChunkMetrics(resp *llm.CompletionResponse, remainingIn, remainingOut int, remainingCost float64, remainingWeight int, taskTokens int, last bool) (int, int, float64) {
	if last {
		return remainingIn, remainingOut, remainingCost
	}

	weight := max(taskTokens, 1)
	if remainingWeight <= 0 {
		remainingWeight = weight
	}
	input := int(float64(resp.TokensInput) * float64(weight) / float64(remainingWeight))
	output := int(float64(resp.TokensOutput) * float64(weight) / float64(remainingWeight))
	cost := resp.EstimatedCostUSD * float64(weight) / float64(remainingWeight)
	if input > remainingIn {
		input = remainingIn
	}
	if output > remainingOut {
		output = remainingOut
	}
	if cost > remainingCost {
		cost = remainingCost
	}
	return input, output, cost
}

// extractCodeFromResponse extracts code blocks from LLM response
func extractCodeFromResponse(response string, language string) string {
	// Try to extract from markdown code blocks
	patterns := []string{
		"```" + language + `\n([\s\S]*?)` + "```",
		"```" + `\n([\s\S]*?)` + "```",
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(response); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// If no code blocks, return the whole response (might be plain code)
	return strings.TrimSpace(response)
}

func (e *Engine) postProcess(code string, adapter adapters.LanguageAdapter, language string, ast *models.AST) string {
	// Add standard imports based on language
	var imports string

	switch language {
	case "go":
		imports = `package ` + ast.Package + `_test

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

`
	case "python":
		imports = `import pytest
from unittest.mock import Mock, patch

`
	case "javascript", "typescript":
		// Imports depend on the source file
		imports = ""
	case "rust":
		imports = `#[cfg(test)]
mod tests {
    use super::*;

`
	}

	// For Go, check if package declaration exists
	if language == "go" && strings.Contains(code, "package ") {
		return code
	}

	return imports + code
}

func (e *Engine) writeTestFile(path string, content string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// GetUsage returns LLM usage metrics
func (e *Engine) GetUsage() *llm.UsageMetrics {
	e.usageMu.Lock()
	defer e.usageMu.Unlock()

	usage := e.usage.Clone()
	hits, misses := e.cache.Counts()
	usage.CacheHits = hits
	usage.CacheMisses = misses
	return usage
}

// GetProviderName returns the underlying provider name.
func (e *Engine) GetProviderName() string {
	if e == nil || e.provider == nil {
		return ""
	}
	return e.provider.Name()
}

// GetCacheStats returns cache statistics
func (e *Engine) GetCacheStats() (size int, hits int, misses int, hitRate float64) {
	return e.cache.Stats()
}

func (e *Engine) recordCachedUsage(resp *llm.CompletionResponse) {
	if resp == nil {
		return
	}
	e.usageMu.Lock()
	defer e.usageMu.Unlock()
	if e.usage.Provider == "" {
		e.usage.Provider = firstNonEmpty(resp.Provider, e.provider.Name())
	}
	if e.usage.Model == "" {
		e.usage.Model = resp.Model
	}
	e.usage.CachedTokens += resp.TokensInput
}

func (e *Engine) recordLiveUsage(resp *llm.CompletionResponse, chunkSize int) {
	if resp == nil {
		return
	}
	e.usageMu.Lock()
	defer e.usageMu.Unlock()
	e.usage.Provider = firstNonEmpty(resp.Provider, e.usage.Provider, e.provider.Name())
	e.usage.Model = firstNonEmpty(resp.Model, e.usage.Model)
	e.usage.TotalRequests++
	e.usage.TotalTokensIn += resp.TokensInput
	e.usage.TotalTokensOut += resp.TokensOutput
	e.usage.EstimatedCostUSD += resp.EstimatedCostUSD
	e.usage.ChunkCount++
	if chunkSize > 1 {
		e.usage.BatchCount++
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// GeneratedTestJSON represents the expected JSON structure from LLM
type GeneratedTestJSON struct {
	TestName     string   `json:"test_name"`
	TestCode     string   `json:"test_code"`
	Imports      []string `json:"imports"`
	EdgeCases    []string `json:"edge_cases_covered"`
	Dependencies []string `json:"mocked_dependencies"`
}

// parseStructuredOutput attempts to parse structured JSON from LLM response
func parseStructuredOutput(response string) (*GeneratedTestJSON, error) {
	// Try to find JSON in response
	jsonRegex := regexp.MustCompile(`\{[\s\S]*\}`)
	jsonMatch := jsonRegex.FindString(response)
	if jsonMatch == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result GeneratedTestJSON
	if err := json.Unmarshal([]byte(jsonMatch), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}
