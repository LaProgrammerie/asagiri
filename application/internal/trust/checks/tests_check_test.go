package checks

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
)

func TestTestsRunnerNoCandidatesCapsConfidence(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	appDir := filepath.Join(repo, "application")
	require.NoError(t, os.MkdirAll(appDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "go.mod"), []byte("module fixture.test\n\ngo 1.22\n"), 0o644))
	deps := testDeps(t, repo)
	deps.Investigate = func(ctx context.Context, repoRoot, feature, taskID string, cfg *config.Config) (investigation.InvestigationResult, error) {
		return investigation.InvestigationResult{}, nil
	}
	scope := Scope{
		TrustID:   "trust-test",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
	result, err := TestsRunner{}.Run(context.Background(), scope, deps)
	require.NoError(t, err)
	require.Equal(t, statusWarn, result.Status)
	require.LessOrEqual(t, result.Confidence, testsNoCandidatesCap)
	require.True(t, countSeverity(result.Findings, "warning") > 0)
}

func TestTestsRunnerSkippedWithoutModule(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	deps := testDeps(t, repo)
	scope := Scope{
		TrustID:   "trust-test",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
	result, err := TestsRunner{}.Run(context.Background(), scope, deps)
	require.NoError(t, err)
	require.Equal(t, statusSkipped, result.Status)
}
