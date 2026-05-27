package investigation

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRunInvestigationProducesReport(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, writeMinimalConfig(repo))
	cfg, err := config.Load(filepath.Join(repo, ".asagiri", "config.yaml"), repo)
	require.NoError(t, err)

	req := Request{
		Symptom:  "onboarding fails after invite",
		Feature:  "workspace-saas",
		Flow:     "onboarding",
		Depth:    DepthStandard,
		NoCloud:  true,
		RepoRoot: repo,
	}
	rep, err := RunInvestigation(context.Background(), req, cfg)
	require.NoError(t, err)
	require.NotEmpty(t, rep.ID)
	require.NotEmpty(t, rep.Hypotheses)
	require.FileExists(t, filepath.Join(repo, ".asagiri", "investigations", rep.ID, "report.md"))
	require.FileExists(t, filepath.Join(repo, ".asagiri", "investigations", rep.ID, "context-pack.json"))
	require.FileExists(t, filepath.Join(repo, ".asagiri", "investigations", rep.ID, "context-pack.md"))
	require.FileExists(t, filepath.Join(repo, ".asagiri", "investigations", rep.ID, "replay-pack.json"))
}

func TestResolveScopeInvite(t *testing.T) {
	scope := ResolveScope(Request{Symptom: "onboarding fails after invite", Flow: "onboarding"})
	require.Equal(t, "onboarding", scope.Flow)
	require.Equal(t, "invite_member", scope.Action)
}

func writeMinimalConfig(repo string) error {
	dir := filepath.Join(repo, ".asagiri")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`project:
  name: test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
mcp:
  investigation:
    max_grep_results: 50
    large_file_bytes: 50000
`), 0o644)
}
