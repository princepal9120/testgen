package generator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/adapters"
	"github.com/princepal9120/testgen-cli/internal/llm"
	"github.com/princepal9120/testgen-cli/pkg/models"
)

type generationTask struct {
	id              string
	def             *models.Definition
	testType        string
	prompt          string
	cacheKey        string
	estimatedTokens int
}

type generationChunk struct {
	tasks           []generationTask
	prompt          string
	systemRole      string
	estimatedTokens int
}

type batchedTestsEnvelope struct {
	Tests []batchedTest `json:"tests"`
}

type batchedTest struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Code string `json:"code"`
}

func buildGenerationTasks(provider llm.Provider, config EngineConfig, adapter adapters.LanguageAdapter, definitions []*models.Definition, packageName string) []generationTask {
	tasks := make([]generationTask, 0, len(definitions)*maxInt(1, len(config.TestTypes)))
	model := llm.ResolveModel(config.Provider, "")
	language := adapter.GetLanguage()

	for _, def := range definitions {
		if def == nil {
			continue
		}
		for _, testType := range config.TestTypes {
			promptTemplate := adapter.GetPromptTemplate(testType)
			prompt := fmt.Sprintf(promptTemplate, def.Body, packageName)
			fingerprint := stableDefinitionFingerprint(config.Provider, model, language, packageName, testType, config.Framework, def)
			tasks = append(tasks, generationTask{
				id:              strconv.Itoa(len(tasks) + 1),
				def:             def,
				testType:        testType,
				prompt:          prompt,
				cacheKey:        llm.NewCache(1).GenerateKeyParts(fingerprint),
				estimatedTokens: provider.CountTokens(prompt),
			})
		}
	}

	return tasks
}

func stableDefinitionFingerprint(provider, model, language, packageName, testType, framework string, def *models.Definition) string {
	if def == nil {
		return ""
	}
	parts := []string{
		llm.ResolveProvider(provider),
		llm.ResolveModel(provider, model),
		strings.TrimSpace(language),
		strings.TrimSpace(packageName),
		strings.TrimSpace(testType),
		strings.TrimSpace(framework),
		strings.TrimSpace(def.Name),
		normalizeWhitespace(def.Signature),
		normalizeWhitespace(def.Body),
	}
	return strings.Join(parts, "|")
}

func planGenerationChunks(provider llm.Provider, config EngineConfig, adapter adapters.LanguageAdapter, packageName string, tasks []generationTask) []generationChunk {
	if len(tasks) == 0 {
		return nil
	}

	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	systemRole := defaultSystemRole(adapter)
	maxChunkTokens := maxInt(batchSize*1200, 1200)
	chunks := make([]generationChunk, 0, len(tasks))
	current := make([]generationTask, 0, batchSize)
	currentTokens := 0

	flush := func() {
		if len(current) == 0 {
			return
		}
		chunks = append(chunks, generationChunk{
			tasks:           append([]generationTask(nil), current...),
			systemRole:      systemRole,
			prompt:          buildChunkPrompt(adapter, config.Framework, packageName, current),
			estimatedTokens: currentTokens,
		})
		current = current[:0]
		currentTokens = 0
	}

	for _, task := range tasks {
		taskTokens := maxInt(task.estimatedTokens, 1)
		if len(current) > 0 && (len(current) >= batchSize || currentTokens+taskTokens > maxChunkTokens) {
			flush()
		}
		current = append(current, task)
		currentTokens += taskTokens
	}
	flush()

	return chunks
}

func buildChunkPrompt(adapter adapters.LanguageAdapter, framework, packageName string, tasks []generationTask) string {
	if len(tasks) == 1 {
		return tasks[0].prompt
	}

	var b strings.Builder
	b.WriteString("Generate production-quality ")
	b.WriteString(adapter.GetLanguage())
	b.WriteString(" tests for each target below.\n")
	if strings.TrimSpace(packageName) != "" {
		b.WriteString("Package/module context: ")
		b.WriteString(packageName)
		b.WriteString("\n")
	}
	if strings.TrimSpace(framework) != "" {
		b.WriteString("Preferred framework: ")
		b.WriteString(framework)
		b.WriteString("\n")
	}
	b.WriteString("Return ONLY valid JSON in this exact shape:\n")
	b.WriteString(`{"tests":[{"id":"1","name":"functionName","code":"<test code only>"}]}`)
	b.WriteString("\nDo not wrap the JSON in markdown code fences.\n")
	b.WriteString("Each id must match the corresponding TARGET block.\n\n")

	for _, task := range tasks {
		b.WriteString("TARGET ")
		b.WriteString(task.id)
		b.WriteString(" | ")
		b.WriteString(task.def.Name)
		b.WriteString(" | ")
		b.WriteString(task.testType)
		b.WriteString("\n")
		b.WriteString(task.prompt)
		b.WriteString("\nEND TARGET ")
		b.WriteString(task.id)
		b.WriteString("\n\n")
	}

	return b.String()
}

func defaultSystemRole(adapter adapters.LanguageAdapter) string {
	return fmt.Sprintf("You are an expert %s developer. Generate production-quality tests that follow best practices. Output only the requested artifact.", adapter.GetLanguage())
}

func parseChunkResponse(content, language string, tasks []generationTask) (map[string]string, error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	if len(tasks) == 1 {
		return map[string]string{tasks[0].id: extractCodeFromResponse(content, language)}, nil
	}

	payload := extractJSONPayload(content)
	if payload == "" {
		return nil, fmt.Errorf("no JSON payload found in batched response")
	}

	results := make(map[string]string, len(tasks))

	var envelope batchedTestsEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err == nil && len(envelope.Tests) > 0 {
		for _, item := range envelope.Tests {
			results[item.ID] = strings.TrimSpace(item.Code)
		}
	}

	if len(results) == 0 {
		var items []batchedTest
		if err := json.Unmarshal([]byte(payload), &items); err == nil {
			for _, item := range items {
				results[item.ID] = strings.TrimSpace(item.Code)
			}
		}
	}

	if len(results) == 0 {
		var object map[string]string
		if err := json.Unmarshal([]byte(payload), &object); err == nil {
			for id, code := range object {
				results[id] = strings.TrimSpace(code)
			}
		}
	}

	missing := make([]string, 0)
	for _, task := range tasks {
		if strings.TrimSpace(results[task.id]) == "" {
			missing = append(missing, task.id)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("batched response missing test code for ids: %s", strings.Join(missing, ", "))
	}

	return results, nil
}

func extractJSONPayload(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	re := regexp.MustCompile(`(?s)(\{.*\}|\[.*\])`)
	match := re.FindString(trimmed)
	return strings.TrimSpace(match)
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
