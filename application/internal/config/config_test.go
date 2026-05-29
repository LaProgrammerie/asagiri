package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadVerificationGates(t *testing.T) {
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
verification:
  default_profile: production
  gates:
    production:
      min_confidence:
        security: 0.85
      required_checks:
        - contracts
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Verification.DefaultProfile != "production" {
		t.Fatalf("default_profile: got %q", cfg.Verification.DefaultProfile)
	}
	if cfg.Verification.Gates["production"].MinConfidence["security"] != 0.85 {
		t.Fatalf("security min: got %v", cfg.Verification.Gates["production"].MinConfidence["security"])
	}
	if len(cfg.Verification.Gates["production"].RequiredChecks) != 1 || cfg.Verification.Gates["production"].RequiredChecks[0] != "contracts" {
		t.Fatalf("required_checks: got %v", cfg.Verification.Gates["production"].RequiredChecks)
	}
}

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

func TestLoadExecutionGraphConfig(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "proj")
	requireDirs(t, repo)
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
project:
  name: graph-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
execution_graph:
  enabled: true
  max_parallel: 3
  default_strategy: cost_aware
  require_checkpoints: true
  stop_on_risk: critical
  allow_parallel_agents: false
  require_isolated_worktrees: true
  gates:
    trust_required_for_high_risk: true
    human_approval_for:
      - migration
  rollback:
    require_strategy_for_high_risk: true
    preserve_failed_worktrees: false
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.ExecutionGraph.IsEnabled() {
		t.Fatal("expected execution graph enabled")
	}
	if cfg.ExecutionGraph.MaxParallel != 3 {
		t.Fatalf("max_parallel: got %d", cfg.ExecutionGraph.MaxParallel)
	}
	if cfg.ExecutionGraph.DefaultStrategy != "cost_aware" {
		t.Fatalf("default_strategy: got %q", cfg.ExecutionGraph.DefaultStrategy)
	}
	if cfg.ExecutionGraph.StopOnRisk != "critical" {
		t.Fatalf("stop_on_risk: got %q", cfg.ExecutionGraph.StopOnRisk)
	}
	if cfg.ExecutionGraph.AllowParallelAgents {
		t.Fatal("expected allow_parallel_agents false")
	}
	if !cfg.ExecutionGraph.Gates.TrustRequiredForHighRisk {
		t.Fatal("expected trust_required_for_high_risk true")
	}
	if len(cfg.ExecutionGraph.Gates.HumanApprovalFor) != 1 || cfg.ExecutionGraph.Gates.HumanApprovalFor[0] != "migration" {
		t.Fatalf("human_approval_for: got %v", cfg.ExecutionGraph.Gates.HumanApprovalFor)
	}
	if !cfg.ExecutionGraph.Rollback.RequireStrategyForHighRisk {
		t.Fatal("expected require_strategy_for_high_risk true")
	}
	if cfg.ExecutionGraph.Rollback.PreserveFailedWorktrees {
		t.Fatal("expected preserve_failed_worktrees false")
	}
}

func TestExecutionGraphDisabledPreservesFalse(t *testing.T) {
	repo := t.TempDir()
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte(`project:
  name: graph-off
state:
  backend: sqlite
  path: .asagiri/state.sqlite
execution_graph:
  enabled: false
`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ExecutionGraph.IsEnabled() {
		t.Fatal("expected execution graph disabled")
	}
}

func TestExecutionGraphDefaults(t *testing.T) {
	cfg := NewTestConfig("proj")
	if !cfg.ExecutionGraph.IsEnabled() {
		t.Fatal("expected execution graph enabled by default")
	}
	if cfg.ExecutionGraph.MaxParallel != 2 {
		t.Fatalf("default max_parallel: got %d", cfg.ExecutionGraph.MaxParallel)
	}
	if cfg.ExecutionGraph.DefaultStrategy != "risk_aware" {
		t.Fatalf("default strategy: got %q", cfg.ExecutionGraph.DefaultStrategy)
	}
	if cfg.ExecutionGraph.StopOnRisk != "high" {
		t.Fatalf("default stop_on_risk: got %q", cfg.ExecutionGraph.StopOnRisk)
	}
}

func TestCoordinationDefaults(t *testing.T) {
	cfg := NewTestConfig("proj")
	if cfg.Coordination.MaxParallelAgents != 2 {
		t.Fatalf("max_parallel_agents: got %d", cfg.Coordination.MaxParallelAgents)
	}
	if cfg.Coordination.DefaultIsolation != "isolated_worktree" {
		t.Fatalf("default_isolation: got %q", cfg.Coordination.DefaultIsolation)
	}
	if len(cfg.Coordination.RequireSecurityReviewFor) != 3 {
		t.Fatalf("require_security_review_for: got %v", cfg.Coordination.RequireSecurityReviewFor)
	}
	if !cfg.Coordination.RequireIndependentReview {
		t.Fatal("expected require_independent_review true by default")
	}
	if cfg.Coordination.AllowSelfReview {
		t.Fatal("expected allow_self_review false by default")
	}
}

func TestLoadCoordinationConfig(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "proj")
	requireDirs(t, repo)
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
project:
  name: coord-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
agents:
  cursor:
    command: cursor-agent
  codex:
    command: codex
coordination:
  max_parallel_agents: 3
  default_isolation: readonly
  require_independent_review: true
  allow_self_review: false
  assignment:
    investigation: local
  profiles:
    cursor-impl:
      agent: cursor
      role: implementer
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath, repo)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Coordination.MaxParallelAgents != 3 {
		t.Fatalf("max_parallel_agents: got %d", cfg.Coordination.MaxParallelAgents)
	}
	if cfg.Coordination.DefaultIsolation != "readonly" {
		t.Fatalf("default_isolation: got %q", cfg.Coordination.DefaultIsolation)
	}
	if cfg.Coordination.Assignment["investigation"] != "local" {
		t.Fatalf("assignment: got %v", cfg.Coordination.Assignment)
	}
	if cfg.Coordination.Profiles["cursor-impl"].Agent != "cursor" {
		t.Fatalf("profile agent: got %q", cfg.Coordination.Profiles["cursor-impl"].Agent)
	}
	if !cfg.Coordination.RequireIndependentReview {
		t.Fatal("expected require_independent_review true")
	}
	if cfg.Coordination.AllowSelfReview {
		t.Fatal("expected allow_self_review false")
	}
}

func TestCoordinationInvalidDefaultIsolationRejected(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "proj")
	requireDirs(t, repo)
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
project:
  name: coord-bad-isolation
state:
  backend: sqlite
  path: .asagiri/state.sqlite
coordination:
  default_isolation: container
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(cfgPath, repo); err == nil {
		t.Fatal("expected validation error for unknown default_isolation")
	}
}

func TestCoordinationProfileUnknownAgentRejected(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "proj")
	requireDirs(t, repo)
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
project:
  name: coord-bad
state:
  backend: sqlite
  path: .asagiri/state.sqlite
agents:
  cursor:
    command: cursor-agent
coordination:
  profiles:
    ghost:
      agent: missing-agent
      role: reviewer
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(cfgPath, repo); err == nil {
		t.Fatal("expected validation error for unknown profile agent")
	}
}

func TestCoordinationHandoffsPathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "proj")
	requireDirs(t, repo)
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
project:
  name: coord-traversal
state:
  backend: sqlite
  path: .asagiri/state.sqlite
coordination:
  handoffs_path: ../../../tmp/evil-handoffs
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(cfgPath, repo); err == nil {
		t.Fatal("expected validation error for handoffs_path traversal")
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

func TestApplyKnowledgeDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.applyKnowledgeDefaults()
	if !cfg.Knowledge.DefaultIncludeFlows || !cfg.Knowledge.DefaultIncludeContracts {
		t.Fatal("expected default include flows and contracts")
	}
	if !cfg.Knowledge.WarnOnStale {
		t.Fatal("expected warn_on_stale default true")
	}
	if cfg.Knowledge.DefaultIncludeCode || cfg.Knowledge.IncrementalByDefault {
		t.Fatal("expected code and incremental defaults false")
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
