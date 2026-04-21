package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// MustLoadFixtureMap loads a JSON fixture relative to baseFile and unmarshals it into a map.
func MustLoadFixtureMap(tb testing.TB, baseFile string, relativePath ...string) map[string]interface{} {
	tb.Helper()

	parts := append([]string{filepath.Dir(baseFile)}, relativePath...)
	path := filepath.Join(parts...)

	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read fixture %s: %v", path, err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		tb.Fatalf("unmarshal fixture %s: %v", path, err)
	}

	return out
}

// AssertJSONSubset verifies that every expected field/value exists in actual.
// Extra fields in actual are allowed so compatibility tests can survive additive schema changes.
func AssertJSONSubset(tb testing.TB, expected, actual interface{}) {
	tb.Helper()
	assertJSONSubset(tb, "$", expected, actual)
}

func assertJSONSubset(tb testing.TB, path string, expected, actual interface{}) {
	tb.Helper()

	switch exp := expected.(type) {
	case map[string]interface{}:
		act, ok := actual.(map[string]interface{})
		if !ok {
			tb.Fatalf("%s: expected object, got %T", path, actual)
		}
		for key, expValue := range exp {
			actValue, ok := act[key]
			if !ok {
				tb.Fatalf("%s.%s: missing key", path, key)
			}
			assertJSONSubset(tb, path+"."+key, expValue, actValue)
		}
	case []interface{}:
		act, ok := actual.([]interface{})
		if !ok {
			tb.Fatalf("%s: expected array, got %T", path, actual)
		}
		if len(act) != len(exp) {
			tb.Fatalf("%s: expected array length %d, got %d", path, len(exp), len(act))
		}
		for idx, expValue := range exp {
			assertJSONSubset(tb, fmt.Sprintf("%s[%d]", path, idx), expValue, act[idx])
		}
	default:
		if !reflect.DeepEqual(expected, actual) {
			tb.Fatalf("%s: expected %#v, got %#v", path, expected, actual)
		}
	}
}

// NormalizeGenerateSuccessPayload replaces dynamic absolute paths so fixtures can stay stable.
func NormalizeGenerateSuccessPayload(payload map[string]interface{}) {
	if payload == nil {
		return
	}

	setPath := func(node map[string]interface{}, key string) {
		if _, ok := node[key].(string); ok {
			node[key] = "__TMP__/sample.py"
		}
	}

	setPath(payload, "target_path")

	if sourceFiles, ok := payload["source_files"].([]interface{}); ok && len(sourceFiles) > 0 {
		if item, ok := sourceFiles[0].(map[string]interface{}); ok {
			setPath(item, "path")
		}
	}

	if results, ok := payload["results"].([]interface{}); ok && len(results) > 0 {
		if result, ok := results[0].(map[string]interface{}); ok {
			if sourceFile, ok := result["source_file"].(map[string]interface{}); ok {
				setPath(sourceFile, "path")
			}
		}
	}
}
