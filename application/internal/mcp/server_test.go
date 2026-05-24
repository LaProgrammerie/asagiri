package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestServerToolReadFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.go")
	if err := os.WriteFile(p, []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewTestConfig("t")
	var in bytes.Buffer
	in.WriteString(`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"asagiri.read_file_safe","arguments":{"path":"a.go"}}}` + "\n")
	var out bytes.Buffer
	s := &Server{RepoRoot: dir, Config: cfg, In: &in, Out: &out}
	if err := s.Serve(context.Background()); err != nil {
		t.Fatal(err)
	}
	var resp rpcResponse
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("%s: %v", out.String(), err)
	}
	if resp.Error != nil {
		t.Fatal(resp.Error.Message)
	}
	resStr := string(mustMarshal(resp.Result))
	if !strings.Contains(resStr, "package x") {
		t.Fatalf("result: %s", resStr)
	}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
