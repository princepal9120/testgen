/*
Package metrics provides usage and cost tracking for TestGen.
*/
package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// RunMetrics represents metrics for a single run
type RunMetrics struct {
	RunID                string    `json:"run_id"`
	Timestamp            time.Time `json:"timestamp"`
	Operation            string    `json:"operation,omitempty"`
	TargetPath           string    `json:"target_path,omitempty"`
	MachineMode          bool      `json:"machine_mode,omitempty"`
	TotalFiles           int       `json:"total_files"`
	TokensInput          int       `json:"tokens_input"`
	TokensOutput         int       `json:"tokens_output"`
	TokensCached         int       `json:"tokens_cached"`
	CacheHitRate         float64   `json:"cache_hit_rate"`
	TotalCostUSD         float64   `json:"total_cost_usd"`
	CoveragePercent      float64   `json:"coverage_percent,omitempty"`
	ExactFunctionFiles   int       `json:"exact_function_files,omitempty"`
	HeuristicFunctionFiles int     `json:"heuristic_function_files,omitempty"`
	MissingTestsCount    int       `json:"missing_tests_count,omitempty"`
	ValidationErrorCount int       `json:"validation_error_count,omitempty"`
	ValidationPassed     bool      `json:"validation_passed,omitempty"`
	ExecutionTimeSeconds float64   `json:"execution_time_seconds"`
	SuccessCount         int       `json:"success_count"`
	ErrorCount           int       `json:"error_count"`
}

// Collector collects and stores metrics
type Collector struct {
	metricsDir string
	current    *RunMetrics
	startTime  time.Time
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	// Use .testgen/metrics in current directory
	metricsDir := filepath.Join(".testgen", "metrics")
	_ = os.MkdirAll(metricsDir, 0755)

	runID := time.Now().Format("20060102-150405")

	return &Collector{
		metricsDir: metricsDir,
		current: &RunMetrics{
			RunID:     runID,
			Timestamp: time.Now(),
		},
		startTime: time.Now(),
	}
}

// SetContext records top-level metadata for the run.
func (c *Collector) SetContext(operation string, targetPath string, machineMode bool) {
	c.current.Operation = operation
	c.current.TargetPath = targetPath
	c.current.MachineMode = machineMode
}

// RecordFile records a file being processed
func (c *Collector) RecordFile(success bool) {
	c.current.TotalFiles++
	if success {
		c.current.SuccessCount++
	} else {
		c.current.ErrorCount++
	}
}

// RecordTokens records token usage
func (c *Collector) RecordTokens(input, output int, cached bool) {
	c.current.TokensInput += input
	c.current.TokensOutput += output
	if cached {
		c.current.TokensCached += input
	}
}

// RecordCost records cost
func (c *Collector) RecordCost(costUSD float64) {
	c.current.TotalCostUSD += costUSD
}

// SetCacheHitRate sets the cache hit rate
func (c *Collector) SetCacheHitRate(rate float64) {
	c.current.CacheHitRate = rate
}

// SetAnalyzeSummary stores trust-related analysis metadata.
func (c *Collector) SetAnalyzeSummary(totalFiles int, exactFiles int, heuristicFiles int) {
	c.current.TotalFiles = totalFiles
	c.current.ExactFunctionFiles = exactFiles
	c.current.HeuristicFunctionFiles = heuristicFiles
	c.current.SuccessCount = totalFiles
}

// SetValidationSummary stores validation trust metrics for a run.
func (c *Collector) SetValidationSummary(totalFiles int, coveragePercent float64, testsPassed int, testsFailed int, missingTests int, validationErrors int) {
	c.current.TotalFiles = totalFiles
	c.current.CoveragePercent = coveragePercent
	c.current.MissingTestsCount = missingTests
	c.current.ValidationErrorCount = validationErrors
	c.current.SuccessCount = testsPassed
	c.current.ErrorCount = testsFailed + validationErrors
	c.current.ValidationPassed = testsFailed == 0 && validationErrors == 0 && missingTests == 0
}

// Finalize completes metrics collection
func (c *Collector) Finalize() *RunMetrics {
	c.current.ExecutionTimeSeconds = time.Since(c.startTime).Seconds()
	return c.current
}

// Save saves metrics to disk
func (c *Collector) Save() error {
	c.Finalize()

	filename := filepath.Join(c.metricsDir, c.current.RunID+".json")

	data, err := json.MarshalIndent(c.current, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetCurrent returns current metrics
func (c *Collector) GetCurrent() *RunMetrics {
	return c.current
}
