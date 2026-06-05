package trust

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

func TestEngineVerifyMinimalProductFixture(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	fixtureRoot := filepath.Join(filepath.Dir(file), "checks", "testdata")

	repo := t.TempDir()
	srcProduct := filepath.Join(fixtureRoot, "minimal-product")
	destProduct := filepath.Join(repo, ".asagiri", "products", "minimal-product")
	require.NoError(t, copyTreeForTest(srcProduct, destProduct))
	analysisDir := filepath.Join(repo, ".asagiri", "analysis", "minimal-product")
	require.NoError(t, os.MkdirAll(analysisDir, 0o755))
	require.NoError(t, copyFileForTest(
		filepath.Join(fixtureRoot, "graphs-minimal.json"),
		filepath.Join(analysisDir, "graphs.json"),
	))
	appDir := filepath.Join(repo, "application")
	require.NoError(t, os.MkdirAll(appDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "go.mod"), []byte("module fixture.test\n\ngo 1.22\n"), 0o644))

	deps := checks.DefaultDependencies()
	deps.Investigate = func(ctx context.Context, repoRoot, feature, taskID string, cfg *config.Config) (investigation.InvestigationResult, error) {
		return investigation.InvestigationResult{
			CandidateFiles: []string{"application/internal/trust/engine.go"},
		}, nil
	}
	deps.ParseFailedTests = func(ctx context.Context, repoRoot string) ([]string, error) {
		return nil, nil
	}
	eng := NewEngineWithChecks(repo, checks.NewDefaultRegistry(deps))

	result, err := eng.Verify(context.Background(), VerificationRequest{
		Flow:    "workspace-onboarding",
		Product: "minimal-product",
	})
	require.NoError(t, err)
	require.Len(t, result.Report.Checks, 15)
	require.Equal(t, CheckStaticAnalysis, result.Report.Checks[0].Type)
	require.Equal(t, CheckType("knowledge-graph"), result.Report.Checks[3].Type)
	require.Equal(t, CheckBlastRadius, result.Report.Checks[13].Type)
	require.Equal(t, CheckType("tests"), result.Report.Checks[14].Type)

	conf := result.Report.Confidence
	require.Greater(t, conf.Overall, 0.0)
	for _, d := range confidence.AllDimensions {
		require.Greater(t, conf.ScoreFor(d), 0.0, "dimension %s", d)
	}
	require.Empty(t, conf.InferredDimensions)
	require.Greater(t, conf.Observability, confidence.InferredDimensionCap)
	require.Greater(t, conf.Security, confidence.InferredDimensionCap)
	require.NotEqual(t, ResidualRiskUnknown, result.Report.ResidualRisk)
	require.NotNil(t, result.Report.BlastRadius)
	require.GreaterOrEqual(t, result.Report.BlastRadius.FlowsImpacted, 1)

	body, err := os.ReadFile(result.JSONPath)
	require.NoError(t, err)
	var decoded TrustReport
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Len(t, decoded.Checks, 15)
	require.NotNil(t, decoded.BlastRadius)

	mdBody, err := os.ReadFile(result.MDPath)
	require.NoError(t, err)
	require.NotContains(t, string(mdBody), ConfidenceUnavailableLabel)
	require.NotContains(t, string(mdBody), ConfidenceInferredCapLabel)
	require.Contains(t, string(mdBody), "## Blast Radius")
}

func copyFileForTest(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	return err
}

func copyTreeForTest(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFileForTest(path, target)
	})
}
