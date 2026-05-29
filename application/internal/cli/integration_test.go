package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func runGitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestCLIIntegrationDryRun(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml.example"), `project:
  name: test
  default_branch: main
specs:
  kiro_path: .kiro/specs
  active_spec_path: docs/ai/active/current-spec.md
  handoff_path: docs/ai/active/handoff.md
state:
  backend: sqlite
  path: .asagiri/state.sqlite
worktrees:
  base_path: .asagiri/worktrees
  branch_prefix: asagiri
  cleanup_policy: keep_failed
agents:
  kiro:
    command: kiro
    args: ["--cli"]
  cursor:
    command: cursor
    args: ["agent", "run"]
  codex:
    command: codex
    args: ["exec"]
  ollama:
    command: ollama
    args: ["run", "qwen2.5-coder:7b"]
  claude:
    command: claude
    args: ["code"]
`)
	writeFile(t, filepath.Join(repo, ".kiro", "specs", "agentflow-test", "requirements.md"), "# Requirements")
	writeFile(t, filepath.Join(repo, ".kiro", "specs", "agentflow-test", "tasks.md"), "- [ ] Build plan\n- [ ] Run dev\n")

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	for _, args := range [][]string{
		{"init"},
		{"prototype", "create", "workspace onboarding", "--product", "workspace-saas"},
		{"flows", "extract", "workspace-saas"},
		{"contracts", "extract", "workspace-saas"},
		{"flows", "review", "workspace-saas"},
		{"architecture", "derive", "workspace-saas"},
		{"spec", "generate-from-product", "workspace-saas"},
		{"product", "review", "workspace-saas"},
		{"flows", "inspect", "workspace-saas"},
		{"flows", "e2e", "workspace-saas", "--flow", "workspace-onboarding", "--dry-run"},
		{"prototype", "run", "workspace-saas", "--dry-run"},
		{"index", "--dry-run"},
		{"work", "développe agentflow-test", "--dry-run", "--plan-only"},
		{"work", "développe agentflow-test", "--dry-run", "--estimate-only"},
		{"daemon", "start"},
		{"daemon", "status"},
		{"session", "create", "onboarding-redesign", "--product", "workspace-saas"},
		{"session", "list"},
		{"runtime", "events"},
		{"investigate", "onboarding fails after invite", "--flow", "onboarding", "--no-cloud", "--estimate-only"},
		{"skills", "list"},
		{"memory", "list"},
		{"work", "develop workspace-saas", "--dry-run", "--plan-only", "--investigate-first"},
		{"estimate", "agentflow-test", "--dry-run"},
		{"inbox", "--source", "local"},
		{"continue", "--dry-run"},
		{"next", "--feature", "agentflow-test"},
		{"plan", "agentflow-test", "--dry-run"},
		{"enrich", "agentflow-test", "--dry-run"},
		{"dev", "agentflow-test", "--dry-run"},
		{"verify", "agentflow-test", "--dry-run"},
		{"review", "agentflow-test", "--dry-run"},
		{"status", "--dry-run"},
	} {
		root.SetArgs(args)
		require.NoError(t, root.Execute(), output.String())
	}

	runID, err := latestRunID(repo)
	require.NoError(t, err)
	require.NotEmpty(t, runID)

	root.SetArgs([]string{"report", runID, "--dry-run"})
	require.NoError(t, root.Execute(), output.String())
	require.FileExists(t, filepath.Join(repo, ".asagiri", "runs", runID, "report.md"))
}

func latestRunID(repo string) (string, error) {
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := loadContext(repo, true)
	if err != nil {
		return "", err
	}
	defer store.Close()
	runs, err := store.Workflow().Status(1)
	if err != nil {
		return "", err
	}
	if len(runs) == 0 {
		return "", nil
	}
	_ = dbPath
	return runs[0].ID, nil
}
