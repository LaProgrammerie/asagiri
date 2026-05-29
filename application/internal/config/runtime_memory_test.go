package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRuntimeMemoryConfig(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "proj")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte(`
project:
  name: test-proj
state:
  backend: sqlite
  path: .asagiri/state.sqlite
runtime:
  memory:
    embedder: ollama
    ollama:
      base_url: http://127.0.0.1:11434
      model: nomic-embed-text
    cloud:
      enabled: false
`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Runtime.Memory.Embedder != "ollama" {
		t.Fatalf("embedder: got %q", cfg.Runtime.Memory.Embedder)
	}
	if cfg.Runtime.Memory.Ollama.Model != "nomic-embed-text" {
		t.Fatalf("ollama model: got %q", cfg.Runtime.Memory.Ollama.Model)
	}
	if cfg.Runtime.Memory.Cloud.Enabled {
		t.Fatal("cloud should stay disabled")
	}
}
