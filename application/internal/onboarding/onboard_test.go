package onboarding_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func runGit(t *testing.T, dir string, args ...string) {
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

func minimalExampleConfig() string {
	return `project:
  name: my-project
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
agents:
  cursor:
    command: echo
    args: ["agent"]
validation:
  commands: []
`
}

func TestOnboardGoFixture(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "t@example.com")
	runGit(t, repo, "config", "user.name", "T")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/t\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml.example"), minimalExampleConfig())
	writeFile(t, filepath.Join(repo, ".gitignore"), ".asagiri/state.sqlite\nworktrees/\n")

	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	require.NoError(t, bootstrap.Init(repo))

	opts := onboarding.Options{Yes: true, NonInteractive: true, Plain: true}
	var out bytes.Buffer
	res, err := onboarding.Onboard(repo, opts, nil, &out)
	require.NoError(t, err)
	require.NotEmpty(t, res.PlannedChanges)

	cfg, err := config.Load(config.ConfigPath(repo), repo)
	require.NoError(t, err)
	require.NotEmpty(t, cfg.Validation.Commands)
	foundGo := false
	for _, c := range cfg.Validation.Commands {
		if c.Command == "go test ./..." {
			foundGo = true
		}
	}
	require.True(t, foundGo)

	// idempotent second run
	res2, err := onboarding.Onboard(repo, opts, nil, &out)
	require.NoError(t, err)
	require.LessOrEqual(t, len(res2.Report.Checks), len(res.Report.Checks)+5)
}

func TestOnboardCastorDryRun(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "t@example.com")
	runGit(t, repo, "config", "user.name", "T")
	writeFile(t, filepath.Join(repo, "castor.php"), "<?php\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml.example"), minimalExampleConfig())

	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	require.NoError(t, bootstrap.Init(repo))

	opts := onboarding.Options{Yes: true, NonInteractive: true, DryRun: true}
	var out bytes.Buffer
	res, err := onboarding.Onboard(repo, opts, nil, &out)
	require.NoError(t, err)
	require.NotEmpty(t, res.PlannedChanges)
	_, err = os.Stat(filepath.Join(repo, ".kiro", "specs"))
	require.True(t, os.IsNotExist(err))
}

func TestReadyPlainGolden(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "t@example.com")
	runGit(t, repo, "config", "user.name", "T")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml.example"), minimalExampleConfig())
	writeFile(t, filepath.Join(repo, ".gitignore"), ".asagiri/state.sqlite\nworktrees/\n")

	oldWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	require.NoError(t, bootstrap.Init(repo))

	var out bytes.Buffer
	_, err := onboarding.Ready(repo, onboarding.Options{Plain: true}, &out)
	require.NoError(t, err)
	got := out.String()
	require.Contains(t, got, "Readiness:")
	require.Contains(t, got, "score")
	require.Contains(t, got, "[")
}

func TestApplyFormCastorIntegration(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "t@example.com")
	runGit(t, repo, "config", "user.name", "T")
	writeFile(t, filepath.Join(repo, "castor.php"), "<?php\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml.example"), minimalExampleConfig())
	writeFile(t, filepath.Join(repo, ".gitignore"), ".asagiri/state.sqlite\n")
	require.NoError(t, bootstrap.Init(repo))

	form := onboarding.BuildForm(repo, onboarding.State{}, nil)
	form.Answers.ProjectName = "chatbot-php"
	form.Answers.DefaultBranch = "main"
	form.Answers.Stack = "castor"
	form.Answers.DefaultAgent = "cursor"
	form.Answers.DefaultReviewer = "codex"
	form.Answers.FeatureSlug = "chatbot-mvp"
	form.Answers.ProductOneLiner = "Chatbot PHP Castor"

	res, err := onboarding.ApplyForm(repo, form)
	require.NoError(t, err)
	require.Greater(t, res.Report.Score, 0)
	require.FileExists(t, filepath.Join(repo, ".asagiri", "config.yaml"))
}
