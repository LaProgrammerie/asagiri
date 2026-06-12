package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func minimalConfigYAML(extra string) string {
	base := `
project:
  name: prov-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
`
	return base + extra
}

func loadConfigYAML(t *testing.T, repo, yamlBody string) *Config {
	t.Helper()
	requireDirs(t, repo)
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlBody), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

func TestLoadLegacyAgentCommandOnly(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfg := loadConfigYAML(t, repo, minimalConfigYAML(`
agents:
  claude:
    command: claude
    args:
      - --print
`))

	a, err := cfg.LookupAgent("claude")
	if err != nil {
		t.Fatal(err)
	}
	if a.Command != "claude" || len(a.Args) != 1 || a.Args[0] != "--print" {
		t.Fatalf("agent: %+v", a)
	}
	typ, err := cfg.AgentProviderType("claude")
	if err != nil {
		t.Fatal(err)
	}
	if typ != ProviderTypeExec {
		t.Fatalf("provider type = %q, want exec", typ)
	}
}

func TestLoadProvidersAndAgentProviderRef(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfg := loadConfigYAML(t, repo, minimalConfigYAML(`
providers:
  kiro-cli:
    type: kiro-cli
    command: kiro
  claude-code:
    type: claude-code
    command: claude
    args:
      - --print
      - --output-format
      - stream-json

agents:
  laprogrammerie:
    provider: kiro-cli
    profile: laprogrammerie
    timeout: 900
  dev:
    provider: claude-code
    model: sonnet
    timeout: 600
`))

	a, err := cfg.LookupAgent("laprogrammerie")
	if err != nil {
		t.Fatal(err)
	}
	if a.Provider != "kiro-cli" || a.Profile != "laprogrammerie" || a.Timeout != 900 {
		t.Fatalf("laprogrammerie: %+v", a)
	}
	typ, err := cfg.AgentProviderType("laprogrammerie")
	if err != nil || typ != ProviderTypeKiroCLI {
		t.Fatalf("type = %q err = %v", typ, err)
	}

	dev, err := cfg.LookupAgent("dev")
	if err != nil {
		t.Fatal(err)
	}
	if dev.Model != "sonnet" {
		t.Fatalf("dev model: %q", dev.Model)
	}
	typ, err = cfg.AgentProviderType("dev")
	if err != nil || typ != ProviderTypeClaudeCode {
		t.Fatalf("dev type = %q err = %v", typ, err)
	}

	p, err := cfg.LookupProvider("claude-code")
	if err != nil {
		t.Fatal(err)
	}
	if p.Command != "claude" || len(p.Args) != 3 {
		t.Fatalf("provider claude-code: %+v", p)
	}
}

func TestLoadAgentUnknownProviderRejected(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	requireDirs(t, repo)
	body := minimalConfigYAML(`
agents:
  dev:
    provider: missing-provider
    model: sonnet
`)
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(cfgPath, repo)
	if err == nil {
		t.Fatal("expected error for unknown provider ref")
	}
	if !strings.Contains(err.Error(), "missing-provider") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadProviderMissingTypeRejected(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	requireDirs(t, repo)
	body := minimalConfigYAML(`
providers:
  broken:
    command: foo
agents:
  dev:
    provider: broken
`)
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(cfgPath, repo)
	if err == nil {
		t.Fatal("expected error for provider without type")
	}
	if !strings.Contains(err.Error(), "providers.broken.type") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadProviderUnknownTypeRejected(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	requireDirs(t, repo)
	body := minimalConfigYAML(`
providers:
  weird:
    type: not-a-real-adapter
    command: foo
`)
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(cfgPath, repo)
	if err == nil {
		t.Fatal("expected error for unknown provider type")
	}
	if !strings.Contains(err.Error(), "not-a-real-adapter") {
		t.Fatalf("error = %v", err)
	}
}

func TestLookupAgentUnknown(t *testing.T) {
	cfg := NewTestConfig("t")
	_, err := cfg.LookupAgent("ghost")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("error = %v", err)
	}
}

func TestLookupProviderUnknown(t *testing.T) {
	cfg := NewTestConfig("t")
	_, err := cfg.LookupProvider("nope")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("error = %v", err)
	}
}

func TestAgentEnvAndProfileRoundTrip(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfg := loadConfigYAML(t, repo, minimalConfigYAML(`
providers:
  ollama:
    type: ollama
    command: ollama
    env:
      OLLAMA_HOST: http://127.0.0.1:11434

agents:
  enrich:
    provider: ollama
    model: qwen2.5-coder:7b
    env:
      OLLAMA_HOST: http://localhost:11434
    timeout: 300
`))

	a, err := cfg.LookupAgent("enrich")
	if err != nil {
		t.Fatal(err)
	}
	if a.Env["OLLAMA_HOST"] != "http://localhost:11434" {
		t.Fatalf("agent env: %+v", a.Env)
	}
}

func TestLegacyWorkDefaultsWithInlineAgents(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfg := loadConfigYAML(t, repo, minimalConfigYAML(`
agents:
  cursor:
    command: cursor-agent
  kiro:
    command: kiro
  codex:
    command: codex
  ollama:
    command: ollama

work:
  default_spec_agent: kiro
  default_agent: cursor
  default_reviewer: codex
  default_enricher: ollama
`))

	for _, name := range []string{
		cfg.Work.DefaultSpecAgent,
		cfg.Work.DefaultAgent,
		cfg.Work.DefaultReviewer,
		cfg.Work.DefaultEnricher,
	} {
		if _, err := cfg.LookupAgent(name); err != nil {
			t.Fatalf("lookup %q: %v", name, err)
		}
	}
	typ, err := cfg.AgentProviderType("cursor")
	if err != nil {
		t.Fatal(err)
	}
	if typ != ProviderTypeExec {
		t.Fatalf("legacy cursor type = %q, want exec", typ)
	}
}

func TestWorkDefaultsStillReferenceAgentNames(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "proj")
	cfg := loadConfigYAML(t, repo, minimalConfigYAML(`
providers:
  kiro-cli:
    type: kiro-cli
    command: kiro

agents:
  laprogrammerie:
    provider: kiro-cli
    profile: laprogrammerie
  cursor:
    command: cursor-agent
  codex:
    command: codex
  ollama:
    command: ollama

work:
  default_spec_agent: laprogrammerie
  default_agent: cursor
  default_reviewer: codex
  default_enricher: ollama
`))

	for _, name := range []string{
		cfg.Work.DefaultSpecAgent,
		cfg.Work.DefaultAgent,
		cfg.Work.DefaultReviewer,
		cfg.Work.DefaultEnricher,
	} {
		if _, err := cfg.LookupAgent(name); err != nil {
			t.Fatalf("work ref %q: %v", name, err)
		}
	}
}

func TestIsKnownProviderType(t *testing.T) {
	for _, typ := range KnownProviderTypes() {
		if !IsKnownProviderType(typ) {
			t.Fatalf("expected known: %q", typ)
		}
	}
	if IsKnownProviderType("by-agent-name") {
		t.Fatal("agent name must not be treated as provider type")
	}
}
