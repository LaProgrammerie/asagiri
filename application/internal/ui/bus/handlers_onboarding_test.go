package bus_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestGetOnboardingWizardQuery(t *testing.T) {
	repo := initOnboardRepo(t)
	qb := bus.NewQueryBus(bus.Deps{RepoRoot: repo})
	res, err := qb.Query(context.Background(), bus.GetOnboardingWizardQuery{})
	require.NoError(t, err)
	typed, ok := res.(bus.OnboardingWizardResult)
	require.True(t, ok)
	require.Equal(t, "welcome", typed.CurrentStep)
	require.NotEmpty(t, typed.Fields["project_name"])
}

func TestAdvanceOnboardingStepCommand(t *testing.T) {
	repo := initOnboardRepo(t)
	cb := bus.NewCommandBus(bus.Deps{RepoRoot: repo})
	fields := map[string]string{
		"project_name":   "chatbot",
		"default_branch": "main",
	}
	_, err := cb.Dispatch(context.Background(), bus.AdvanceOnboardingStepCommand{
		Direction: "next",
		Fields:    fields,
	})
	require.NoError(t, err)
	st, err := onboarding.LoadState(repo)
	require.NoError(t, err)
	require.Equal(t, onboarding.StepProject, st.CurrentStep)

	_, err = cb.Dispatch(context.Background(), bus.AdvanceOnboardingStepCommand{
		Direction: "next",
		Fields:    fields,
	})
	require.NoError(t, err)
	st, err = onboarding.LoadState(repo)
	require.NoError(t, err)
	require.Equal(t, onboarding.StepStack, st.CurrentStep)
}

func TestApplyOnboardingConfigFromWizard(t *testing.T) {
	repo := initCastorRepo(t)
	cb := bus.NewCommandBus(bus.Deps{RepoRoot: repo})
	fields := map[string]string{
		"project_name":      "chatbot-php",
		"default_branch":    "main",
		"stack":             "castor",
		"default_spec_agent": "kiro",
		"default_enricher":   "ollama",
		"default_agent":      "cursor",
		"default_reviewer":   "codex",
		"feature_slug":      "chatbot-mvp",
		"product_one_liner": "Chatbot Castor",
	}
	res, err := cb.Dispatch(context.Background(), bus.ApplyOnboardingConfigCommand{
		Yes:    true,
		Fields: fields,
	})
	require.NoError(t, err)
	require.True(t, res.Accepted)

	cfg, err := config.Load(config.ConfigPath(repo), repo)
	require.NoError(t, err)
	require.Equal(t, "chatbot-php", cfg.Project.Name)
	foundCastor := false
	for _, c := range cfg.Validation.Commands {
		if strings.Contains(c.Command, "castor") {
			foundCastor = true
		}
	}
	require.True(t, foundCastor)
}

func TestGetReadinessQuery(t *testing.T) {
	repo := t.TempDir()
	cfg := config.NewTestConfig("test")
	qb := bus.NewQueryBus(bus.Deps{RepoRoot: repo, Config: cfg})
	res, err := qb.Query(context.Background(), bus.GetReadinessQuery{})
	require.NoError(t, err)
	typed, ok := res.(bus.ReadinessResult)
	require.True(t, ok)
	require.GreaterOrEqual(t, typed.Score, 0)
}

func initOnboardRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "t@example.com")
	runGit(t, repo, "config", "user.name", "T")
	writeOnboardFile(t, filepath.Join(repo, ".asagiri", "config.yaml.example"), minimalExampleConfig())
	require.NoError(t, bootstrap.Init(repo))
	return repo
}

func initCastorRepo(t *testing.T) string {
	t.Helper()
	repo := initOnboardRepo(t)
	writeOnboardFile(t, filepath.Join(repo, "castor.php"), "<?php\n")
	return repo
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func writeOnboardFile(t *testing.T, path, content string) {
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
validation:
  commands: []
`
}
