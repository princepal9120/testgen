package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/princepal9120/testgen-cli/internal/app"
)

const protocolVersion = "2025-03-26"

type Server struct {
	service *app.Service
	version string
	name    string
}

func NewServer(version string) *Server {
	return &Server{
		service: app.NewService(),
		version: version,
		name:    "testgen",
	}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *respError  `json:"error,omitempty"`
}

type respError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)
	for {
		msg, err := readMessage(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var req request
		if err := json.Unmarshal(msg, &req); err != nil {
			if err := writeResponse(out, response{
				JSONRPC: "2.0",
				Error:   &respError{Code: -32700, Message: "invalid JSON"},
			}); err != nil {
				return err
			}
			continue
		}

		if req.Method == "notifications/initialized" || strings.HasPrefix(req.Method, "notifications/") {
			continue
		}

		resp := s.handle(ctx, req)
		if req.ID == nil {
			continue
		}
		if err := writeResponse(out, resp); err != nil {
			return err
		}
	}
}

func (s *Server) handle(ctx context.Context, req request) response {
	id := rawID(req.ID)
	switch req.Method {
	case "initialize":
		return response{
			JSONRPC: "2.0",
			ID:      id,
			Result: map[string]interface{}{
				"protocolVersion": protocolVersion,
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{
						"listChanged": false,
					},
				},
				"serverInfo": map[string]interface{}{
					"name":    s.name,
					"version": s.version,
				},
			},
		}
	case "ping":
		return response{JSONRPC: "2.0", ID: id, Result: map[string]interface{}{}}
	case "tools/list":
		return response{JSONRPC: "2.0", ID: id, Result: map[string]interface{}{"tools": s.tools()}}
	case "tools/call":
		result, err := s.callTool(ctx, req.Params)
		if err != nil {
			return response{JSONRPC: "2.0", ID: id, Error: &respError{Code: -32000, Message: err.Error()}}
		}
		return response{JSONRPC: "2.0", ID: id, Result: result}
	default:
		return response{
			JSONRPC: "2.0",
			ID:      id,
			Error:   &respError{Code: -32601, Message: "method not found"},
		}
	}
}

func (s *Server) tools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "testgen_generate",
			"description": "Generate tests using TestGen's shared application layer",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":       map[string]interface{}{"type": "string"},
					"file":       map[string]interface{}{"type": "string"},
					"recursive":  map[string]interface{}{"type": "boolean"},
					"types":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"dry_run":    map[string]interface{}{"type": "boolean"},
					"validate":   map[string]interface{}{"type": "boolean"},
					"emit_patch": map[string]interface{}{"type": "boolean"},
				},
			},
		},
		{
			"name":        "testgen_analyze",
			"description": "Analyze a codebase for test generation readiness and cost",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":          map[string]interface{}{"type": "string"},
					"recursive":     map[string]interface{}{"type": "boolean"},
					"cost_estimate": map[string]interface{}{"type": "boolean"},
					"detail":        map[string]interface{}{"type": "string"},
				},
			},
		},
		{
			"name":        "testgen_validate",
			"description": "Validate generated or existing tests for a path",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":            map[string]interface{}{"type": "string"},
					"recursive":       map[string]interface{}{"type": "boolean"},
					"min_coverage":    map[string]interface{}{"type": "number"},
					"fail_on_missing": map[string]interface{}{"type": "boolean"},
					"report_gaps":     map[string]interface{}{"type": "boolean"},
				},
			},
		},
	}
}

func (s *Server) callTool(ctx context.Context, raw json.RawMessage) (map[string]interface{}, error) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, fmt.Errorf("invalid tools/call params: %w", err)
	}

	var payload interface{}
	switch params.Name {
	case "testgen_generate":
		req := app.GenerateRequest{
			Path:        stringArg(params.Arguments, "path"),
			File:        stringArg(params.Arguments, "file"),
			Recursive:   boolArg(params.Arguments, "recursive"),
			TestTypes:   stringSliceArg(params.Arguments, "types", []string{"unit"}),
			DryRun:      boolArg(params.Arguments, "dry_run"),
			Validate:    boolArg(params.Arguments, "validate"),
			EmitPatch:   boolArg(params.Arguments, "emit_patch"),
			Parallelism: intArg(params.Arguments, "parallelism", 2),
			BatchSize:   intArg(params.Arguments, "batch_size", 5),
			Provider:    stringArg(params.Arguments, "provider"),
		}
		if !req.DryRun && !boolArg(params.Arguments, "write_files") {
			req.DryRun = true
		}
		resp, err := s.service.Generate(ctx, req)
		if err != nil {
			return nil, err
		}
		payload = resp
	case "testgen_analyze":
		resp, err := s.service.Analyze(ctx, app.AnalyzeRequest{
			Path:         stringArg(params.Arguments, "path"),
			Recursive:    boolArgDefault(params.Arguments, "recursive", true),
			CostEstimate: boolArg(params.Arguments, "cost_estimate"),
			Detail:       stringArgDefault(params.Arguments, "detail", "summary"),
		})
		if err != nil {
			return nil, err
		}
		payload = resp
	case "testgen_validate":
		resp, err := s.service.Validate(ctx, app.ValidateRequest{
			Path:          stringArg(params.Arguments, "path"),
			Recursive:     boolArgDefault(params.Arguments, "recursive", true),
			MinCoverage:   floatArg(params.Arguments, "min_coverage"),
			FailOnMissing: boolArg(params.Arguments, "fail_on_missing"),
			ReportGaps:    boolArg(params.Arguments, "report_gaps"),
		})
		if err != nil {
			return nil, err
		}
		payload = resp
	default:
		return nil, fmt.Errorf("unknown tool: %s", params.Name)
	}

	text, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(text),
			},
		},
		"isError": false,
	}, nil
}

func readMessage(r *bufio.Reader) ([]byte, error) {
	contentLength := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			value := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(line), "content-length:"))
			contentLength, err = strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid content length: %w", err)
			}
		}
	}
	if contentLength <= 0 {
		return nil, io.EOF
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

func writeResponse(w io.Writer, resp response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func rawID(id json.RawMessage) interface{} {
	if id == nil {
		return nil
	}
	var out interface{}
	if err := json.Unmarshal(id, &out); err != nil {
		return string(id)
	}
	return out
}

func stringArg(args map[string]interface{}, key string) string {
	return stringArgDefault(args, key, "")
}

func stringArgDefault(args map[string]interface{}, key string, fallback string) string {
	if value, ok := args[key].(string); ok {
		return value
	}
	return fallback
}

func boolArg(args map[string]interface{}, key string) bool {
	return boolArgDefault(args, key, false)
}

func boolArgDefault(args map[string]interface{}, key string, fallback bool) bool {
	if value, ok := args[key].(bool); ok {
		return value
	}
	return fallback
}

func intArg(args map[string]interface{}, key string, fallback int) int {
	if value, ok := args[key].(float64); ok {
		return int(value)
	}
	return fallback
}

func floatArg(args map[string]interface{}, key string) float64 {
	if value, ok := args[key].(float64); ok {
		return value
	}
	return 0
}

func stringSliceArg(args map[string]interface{}, key string, fallback []string) []string {
	raw, ok := args[key]
	if !ok {
		return fallback
	}
	items, ok := raw.([]interface{})
	if !ok {
		return fallback
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
