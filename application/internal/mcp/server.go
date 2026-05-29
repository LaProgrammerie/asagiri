package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/contextopt"
	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
)

// Server is a minimal stdio MCP JSON-RPC server (specv3 §10) — initialize, tools/list, tools/call only.
type Server struct {
	RepoRoot string
	Config   *config.Config
	In       io.Reader
	Out      io.Writer
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type rpcResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Result  any            `json:"result,omitempty"`
	Error   *rpcError      `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Serve blocks processing one line per JSON-RPC request (newline-delimited).
func (s *Server) Serve(ctx context.Context) error {
	sc := bufio.NewScanner(s.In)
	const maxScan = 12 << 20
	sc.Buffer(make([]byte, 0, 64*1024), maxScan)
	for sc.Scan() {
		line := sc.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeErr(nil, -32700, "parse error")
			continue
		}
		resp := s.handle(ctx, req)
		enc, _ := json.Marshal(resp)
		_, _ = s.Out.Write(append(enc, '\n'))
	}
	return sc.Err()
}

func (s *Server) handle(ctx context.Context, req rpcRequest) rpcResponse {
	base := rpcResponse{JSONRPC: "2.0", ID: decodeID(req.ID)}
	switch req.Method {
	case "initialize":
		return finish(base, map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "asa", "version": "v3"},
		})
	case "tools/list":
		return finish(base, map[string]any{"tools": toolDefs()})
	case "tools/call":
		return s.toolCall(ctx, base, req.Params)
	default:
		base.Error = &rpcError{Code: -32601, Message: "method not found"}
		return base
	}
}

func decodeID(raw json.RawMessage) any {
	if raw == nil {
		return nil
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}
	return nil
}

func finish(base rpcResponse, result any) rpcResponse {
	base.Result = result
	return base
}

func (s *Server) writeErr(id any, code int, msg string) {
	resp := rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
	enc, _ := json.Marshal(resp)
	_, _ = s.Out.Write(append(enc, '\n'))
}

func toolDefs() []map[string]any {
	return []map[string]any{
		{"name": "asagiri.search", "inputSchema": map[string]any{}},
		{"name": "asagiri.read_file_safe", "inputSchema": map[string]any{}},
		{"name": "asagiri.extract_symbols", "inputSchema": map[string]any{}},
		{"name": "asagiri.find_related_tests", "inputSchema": map[string]any{}},
		{"name": "asagiri.estimate_tokens", "inputSchema": map[string]any{}},
		{"name": "asagiri.estimate_cost", "inputSchema": map[string]any{}},
		{"name": "asagiri.get_task_context", "inputSchema": map[string]any{}},
		{"name": "asagiri.get_run_status", "inputSchema": map[string]any{}},
		{"name": "asagiri.get_diff_summary", "inputSchema": map[string]any{}},
		{"name": "asagiri.run_local_check", "inputSchema": map[string]any{}},
	}
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) toolCall(ctx context.Context, base rpcResponse, params json.RawMessage) rpcResponse {
	if s.Config == nil {
		base.Error = &rpcError{Code: -32002, Message: "config required"}
		return base
	}
	var p toolCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		base.Error = &rpcError{Code: -32602, Message: "invalid params"}
		return base
	}
	maxB := 256 * 1024
	if s.Config != nil {
		maxB = s.Config.MCP.MaxOutputBytes
	}
	sec := 60
	if s.Config != nil && s.Config.MCP.Investigation.CommandTimeoutSec > 0 {
		sec = s.Config.MCP.Investigation.CommandTimeoutSec
	}
	timeout := time.Duration(sec) * time.Second

	args := map[string]json.RawMessage{}
	if len(p.Arguments) > 0 {
		_ = json.Unmarshal(p.Arguments, &args)
	}

	var text string
	var err error
	switch p.Name {
	case "asagiri.search":
		q := stringArg(args, "query")
		if q == "" {
			err = fmt.Errorf("query requis")
			break
		}
		hits, e := investigation.Grep(ctx, s.RepoRoot, q, s.Config.MCP.Investigation)
		err = e
		text = strings.Join(hits, "\n")
	case "asagiri.read_file_safe":
		rel := stringArg(args, "path")
		if err := s.denyPath(rel); err != nil {
			base.Error = &rpcError{Code: -32003, Message: err.Error()}
			return base
		}
		b, e := investigation.ReadFileSnippet(s.RepoRoot, rel, maxB)
		err = e
		text = string(b)
	case "asagiri.extract_symbols":
		rel := stringArg(args, "path")
		if err := s.denyPath(rel); err != nil {
			base.Error = &rpcError{Code: -32003, Message: err.Error()}
			return base
		}
		b, e := investigation.ReadFileSnippet(s.RepoRoot, rel, maxB)
		err = e
		syms := investigation.ExtractGoSymbols(string(b))
		enc, _ := json.Marshal(syms)
		text = string(enc)
	case "asagiri.find_related_tests":
		var files []string
		_ = json.Unmarshal(args["files"], &files)
		text = strings.Join(investigation.RelatedTestPaths(files), "\n")
	case "asagiri.estimate_tokens":
		content := stringArg(args, "content")
		k := cost.ContentDefault
		if stringArg(args, "kind") == "code" {
			k = cost.ContentCode
		}
		model := stringArg(args, "model")
		var tok int
		if model != "" {
			tok = cost.EstimateFromTextForModel(content, model, k, s.Config.TokenEst)
		} else {
			tok = cost.EstimateFromText(content, k, s.Config.TokenEst)
		}
		text = fmt.Sprintf("%d", tok)
	case "asagiri.estimate_cost":
		model := stringArg(args, "model")
		inTok := intArg(args, "input_tokens")
		outTok := intArg(args, "output_tokens")
		mny, e := cost.CostFromPricing(s.Config, model, inTok, outTok)
		err = e
		text = fmt.Sprintf(`{"cents":%d,"currency":%q}`, mny.Cents, mny.Currency)
	case "asagiri.get_task_context":
		feat := stringArg(args, "feature")
		entries, e := contextopt.Collect(s.RepoRoot, feat, s.Config, contextopt.CollectOpts{MaxFiles: 80})
		err = e
		var lines []string
		for _, e := range entries {
			lines = append(lines, e.RelPath)
		}
		text = strings.Join(lines, "\n")
	case "asagiri.get_run_status":
		text = "not_persisted_in_mcp_stub"
	case "asagiri.get_diff_summary":
		out, e := investigation.RunCommand(ctx, timeout, "git", "-C", s.RepoRoot, "diff", "--stat")
		err = e
		text = string(out)
	case "asagiri.run_local_check":
		out, e := investigation.RunCommand(ctx, timeout, "go", "test", "./...")
		err = e
		text = string(out)
	default:
		base.Error = &rpcError{Code: -32601, Message: "unknown tool"}
		return base
	}
	if err != nil {
		base.Error = &rpcError{Code: -32000, Message: err.Error()}
		return base
	}
	if len(text) > maxB {
		text = text[:maxB] + "\n… truncated …"
	}
	return finish(base, map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
	})
}

func stringArg(args map[string]json.RawMessage, k string) string {
	raw, ok := args[k]
	if !ok {
		return ""
	}
	var s string
	_ = json.Unmarshal(raw, &s)
	return strings.TrimSpace(s)
}

func intArg(args map[string]json.RawMessage, k string) int {
	raw, ok := args[k]
	if !ok {
		return 0
	}
	var n int
	_ = json.Unmarshal(raw, &n)
	return n
}

func (s *Server) denyPath(rel string) error {
	clean := filepath.ToSlash(filepath.Clean(rel))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("path interdit")
	}
	base := filepath.ToSlash(s.RepoRoot)
	abs := filepath.Join(s.RepoRoot, clean)
	if !strings.HasPrefix(filepath.Clean(abs), filepath.Clean(base)) {
		return fmt.Errorf("hors scope repo")
	}
	if s.Config != nil {
		for _, d := range s.Config.MCP.SecretPathDenylist {
			if d != "" && strings.Contains(strings.ToLower(clean), strings.ToLower(d)) {
				return fmt.Errorf("chemin sensible interdit: %s", d)
			}
		}
	}
	return nil
}
