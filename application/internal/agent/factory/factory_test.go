package factory_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agent/claudecode"
	"github.com/LaProgrammerie/asagiri/application/internal/agent/exec"
	"github.com/LaProgrammerie/asagiri/application/internal/agent/factory"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func testConfigWithProviders(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.NewTestConfig("t")
	cfg.Providers = map[string]config.ProviderConfig{
		"claude-code": {
			Type:    config.ProviderTypeClaudeCode,
			Command: "claude",
			Args:    []string{"--print", "--output-format", "stream-json"},
			Env:     map[string]string{"BASE": "provider"},
			Timeout: 120,
		},
		"kiro-cli": {
			Type:    config.ProviderTypeKiroCLI,
			Command: "kiro",
			Args:    []string{"--cli"},
		},
		"ollama": {
			Type:    config.ProviderTypeOllama,
			Command: "ollama",
			Args:    []string{"run"},
		},
	}
	cfg.Agents = map[string]config.Agent{
		"legacy-claude": {
			Command: "claude",
			Args:    []string{"--print"},
		},
		"dev": {
			Provider: "claude-code",
			Model:    "sonnet",
			Args:     []string{"--model", "sonnet"},
			Env:      map[string]string{"BASE": "agent"},
			Timeout:  600,
		},
		"laprogrammerie": {
			Provider: "kiro-cli",
			Profile:  "laprogrammerie",
		},
		"enrich": {
			Provider: "ollama",
			Model:    "qwen2.5-coder:7b",
			Args:     []string{"qwen2.5-coder:7b"},
		},
	}
	return cfg
}

func TestNewFromConfigLegacyUsesExec(t *testing.T) {
	cfg := testConfigWithProviders(t)
	a, err := factory.NewFromConfig("legacy-claude", cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.(*exec.Executor); !ok {
		t.Fatalf("want *exec.Executor, got %T", a)
	}
	if a.Name() != "legacy-claude" {
		t.Fatalf("name = %q", a.Name())
	}
}

func TestNewFromConfigClaudeCodeUsesAdapter(t *testing.T) {
	cfg := testConfigWithProviders(t)
	a, err := factory.NewFromConfig("dev", cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.(*claudecode.Adapter); !ok {
		t.Fatalf("want *claudecode.Adapter, got %T", a)
	}
	if a.Name() != "dev" {
		t.Fatalf("name = %q", a.Name())
	}
}

func TestNewFromConfigKiroCLIUsesExec(t *testing.T) {
	cfg := testConfigWithProviders(t)
	a, err := factory.NewFromConfig("laprogrammerie", cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.(*exec.Executor); !ok {
		t.Fatalf("want *exec.Executor, got %T", a)
	}
}

func TestNewFromConfigOllamaUsesExec(t *testing.T) {
	cfg := testConfigWithProviders(t)
	a, err := factory.NewFromConfig("enrich", cfg, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.(*exec.Executor); !ok {
		t.Fatalf("want *exec.Executor, got %T", a)
	}
}

func TestNewFromConfigUnknownAgent(t *testing.T) {
	cfg := testConfigWithProviders(t)
	_, err := factory.NewFromConfig("missing", cfg, true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMergedAgentRuntimeMergeRules(t *testing.T) {
	cfg := testConfigWithProviders(t)
	typ, merged, err := cfg.MergedAgentRuntime("dev")
	if err != nil {
		t.Fatal(err)
	}
	if typ != config.ProviderTypeClaudeCode {
		t.Fatalf("type = %q", typ)
	}
	if merged.Command != "claude" {
		t.Fatalf("command = %q", merged.Command)
	}
	wantArgs := []string{"--print", "--output-format", "stream-json", "--model", "sonnet"}
	if len(merged.Args) != len(wantArgs) {
		t.Fatalf("args = %v", merged.Args)
	}
	for i, want := range wantArgs {
		if merged.Args[i] != want {
			t.Fatalf("args[%d] = %q want %q (full %v)", i, merged.Args[i], want, merged.Args)
		}
	}
	if merged.Env["BASE"] != "agent" {
		t.Fatalf("env BASE = %q", merged.Env["BASE"])
	}
	if merged.Timeout != 600 {
		t.Fatalf("timeout = %d", merged.Timeout)
	}
}

func TestMergedAgentRuntimeProviderTimeoutFallback(t *testing.T) {
	cfg := testConfigWithProviders(t)
	_, merged, err := cfg.MergedAgentRuntime("laprogrammerie")
	if err != nil {
		t.Fatal(err)
	}
	if merged.Command != "kiro" {
		t.Fatalf("command = %q", merged.Command)
	}
	if len(merged.Args) != 1 || merged.Args[0] != "--cli" {
		t.Fatalf("args = %v", merged.Args)
	}
}
