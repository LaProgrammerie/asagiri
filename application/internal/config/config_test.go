package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
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
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Project.Name != "test-proj" {
		t.Fatalf("name: got %q", cfg.Project.Name)
	}
	if cfg.State.Backend != "sqlite" {
		t.Fatalf("backend: got %q", cfg.State.Backend)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(":\n\tbad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(cfgPath, dir); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidateMCPWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	cfg := NewTestConfig("x")
	cfg.MCP.Enabled = true
	cfg.MCP.MaxOutputBytes = 0
	if err := cfg.Validate(dir); err == nil {
		t.Fatal("expected mcp validation error")
	}
	cfg.MCP.MaxOutputBytes = 1024
	cfg.MCP.CommandTimeoutSec = 30
	cfg.MCP.Investigation.CommandTimeoutSec = 30
	if err := cfg.Validate(dir); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAbsolutePathRejected(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		State: State{Backend: "sqlite", Path: "/tmp/state.sqlite"},
	}
	cfg.applyDefaults("x")
	if err := cfg.Validate(dir); err == nil {
		t.Fatal("expected error for absolute path")
	}
}

func TestValidateEscapeRejected(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		State: State{Backend: "sqlite", Path: "../outside/db.sqlite"},
	}
	cfg.applyDefaults("x")
	if err := cfg.Validate(dir); err == nil {
		t.Fatal("expected error for path escaping repo")
	}
}

func TestLoadPoliciesAndValidation(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "proj")
	requireDirs(t, repo)
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module x\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
project:
  name: test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
validation:
  commands:
    - name: tests
      command: go test ./...
      required: true
policies:
  require_clean_git: true
  max_files_changed_per_task: 15
`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Policies.RequireCleanGit || cfg.Policies.MaxFilesChangedPerTask != 15 {
		t.Fatalf("policies: %+v", cfg.Policies)
	}
	if len(cfg.Validation.Commands) != 1 {
		t.Fatalf("validation: %+v", cfg.Validation.Commands)
	}
}

func requireDirs(t *testing.T, repo string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestValidateUnsupportedBackend(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		State: State{Backend: "postgres", Path: ".asagiri/state.sqlite"},
	}
	cfg.applyDefaults("x")
	if err := cfg.Validate(dir); err == nil {
		t.Fatal("expected backend error")
	}
}
