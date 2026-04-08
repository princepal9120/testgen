package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestServerInitializeAndListTools(t *testing.T) {
	t.Parallel()

	server := NewServer("test")
	in := bytes.NewBuffer(nil)
	out := bytes.NewBuffer(nil)

	writeRequest(in, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]interface{}{},
	})
	writeRequest(in, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	})

	if err := server.Run(context.Background(), in, out); err != nil {
		t.Fatalf("server run failed: %v", err)
	}

	responses := decodeResponses(t, out.Bytes())
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}

	first := responses[0]
	result, ok := first["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected initialize result, got %#v", first)
	}
	if result["protocolVersion"] != protocolVersion {
		t.Fatalf("unexpected protocol version: %#v", result["protocolVersion"])
	}

	second := responses[1]
	tools := second["result"].(map[string]interface{})["tools"].([]interface{})
	if len(tools) < 3 {
		t.Fatalf("expected at least 3 tools, got %d", len(tools))
	}
}

func TestServerToolCallGenerate(t *testing.T) {
	t.Parallel()

	server := NewServer("test")
	dir := t.TempDir()
	file := dir + "/sample.py"
	if err := os.WriteFile(file, []byte("# no defs\n"), 0o644); err != nil {
		t.Fatalf("write sample file: %v", err)
	}

	in := bytes.NewBuffer(nil)
	out := bytes.NewBuffer(nil)
	writeRequest(in, map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "testgen_generate",
			"arguments": map[string]interface{}{
				"file":       file,
				"dry_run":    true,
				"emit_patch": true,
				"types":      []string{"unit"},
			},
		},
	})

	if err := server.Run(context.Background(), in, out); err != nil {
		t.Fatalf("server run failed: %v", err)
	}

	responses := decodeResponses(t, out.Bytes())
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	result := responses[0]["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "\"results\"") {
		t.Fatalf("expected generate payload in tool result, got %s", text)
	}
}

func writeRequest(buf *bytes.Buffer, payload map[string]interface{}) {
	data, _ := json.Marshal(payload)
	buf.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data)))
	buf.Write(data)
}

func decodeResponses(t *testing.T, raw []byte) []map[string]interface{} {
	t.Helper()

	text := string(raw)
	parts := strings.Split(text, "Content-Length:")
	out := make([]map[string]interface{}, 0)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, "\r\n\r\n")
		if idx < 0 {
			t.Fatalf("invalid response framing: %q", part)
		}
		body := part[idx+4:]
		start := strings.Index(body, "{")
		if start < 0 {
			t.Fatalf("missing json body in %q", body)
		}
		body = body[start:]
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			t.Fatalf("invalid response json: %v", err)
		}
		out = append(out, payload)
	}
	return out
}
