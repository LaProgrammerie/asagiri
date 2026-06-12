package trust

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/replay"
)

func TestResolveScope(t *testing.T) {
	repo := t.TempDir()
	productDir := filepath.Join(repo, ".asagiri", "products", "demo")
	require.NoError(t, os.MkdirAll(filepath.Join(productDir, "flows"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(productDir, "flows", "onboarding.flow.yaml"), []byte("id: onboarding\nentry_screen: s1\nsteps:\n  - id: s1\n    screen: s1\n    action: go\n"), 0o644))

	scope, err := ResolveScope(repo, "trust-2026-05-29-abc", VerificationRequest{
		Flow:   "onboarding",
		Task:   "task-1",
		Branch: "feature/x",
	})
	require.NoError(t, err)
	require.Equal(t, "trust-2026-05-29-abc", scope.TrustID)
	require.Equal(t, "onboarding", scope.Flow)
	require.Equal(t, "task-1", scope.Task)
	require.Equal(t, "feature/x", scope.Branch)
	require.Equal(t, repo, scope.RepoRoot)
	require.Equal(t, "demo", scope.ProductID)
}

func TestNewTrustID(t *testing.T) {
	id := NewTrustID()
	require.Regexp(t, `^trust-\d{4}-\d{2}-\d{2}-[a-f0-9]{8}$`, id)
}

func TestNewTrustIDUnique(t *testing.T) {
	a := NewTrustID()
	b := NewTrustID()
	require.NotEqual(t, a, b)
}

func TestEngineVerifyRequiresRepoRoot(t *testing.T) {
	eng := NewEngine("")
	_, err := eng.Verify(context.Background(), VerificationRequest{Flow: "f", Product: "demo"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "repo root required")
}

func TestMapRegistryChecks(t *testing.T) {
	require.Empty(t, mapRegistryChecks(nil))
	require.NotNil(t, mapRegistryChecks(nil))
	out := mapRegistryChecks([]checks.Check{
		{
			ID:         "chk-1",
			Name:       "Static",
			Type:       "static-analysis",
			Status:     "passed",
			Confidence: 0.9,
		},
	})
	require.Len(t, out, 1)
	require.Equal(t, "chk-1", out[0].ID)
	require.Equal(t, CheckStaticAnalysis, out[0].Type)
	require.Equal(t, CheckStatusPassed, out[0].Status)
	require.Equal(t, 0.9, out[0].Confidence)
}

type stubEmitter struct {
	events []string
}

func (s *stubEmitter) Emit(_ context.Context, name string, _ map[string]any) error {
	s.events = append(s.events, name)
	return nil
}

func TestEngineVerifyEmitsLifecycleEvents(t *testing.T) {
	repo := t.TempDir()
	emit := &stubEmitter{}
	eng := NewEngineWithChecks(repo, checks.NewRegistry())
	eng.Emitter = emit

	_, err := eng.Verify(context.Background(), VerificationRequest{Flow: "f", Product: "demo"})
	require.NoError(t, err)
	require.Contains(t, emit.events, "verification.started")
	require.Contains(t, emit.events, "verification.completed")
}

func TestEngineVerify(t *testing.T) {
	repo := t.TempDir()
	eng := NewEngineWithChecks(repo, checks.NewRegistry())

	result, err := eng.Verify(context.Background(), VerificationRequest{
		Flow:    "onboarding-flow",
		Branch:  "onboarding-enterprise",
		Product: "demo",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.TrustID)
	require.FileExists(t, result.MDPath)
	require.FileExists(t, result.JSONPath)

	replayPath := filepath.Join(repo, ".asagiri", "trust", result.TrustID, "replay.yaml")
	require.FileExists(t, replayPath)
	loaded, err := replay.Load(repo, result.TrustID)
	require.NoError(t, err)
	require.Equal(t, "onboarding-flow", loaded.Flow)
	require.Equal(t, "onboarding-enterprise", loaded.Branch)

	require.Equal(t, GateStatusNotConfigured, result.Report.Gate.Status)
	require.Equal(t, ResidualRiskUnknown, result.Report.ResidualRisk)
	require.Zero(t, result.Report.Confidence.Overall)

	body, err := os.ReadFile(result.JSONPath)
	require.NoError(t, err)
	var decoded TrustReport
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, result.TrustID, decoded.TrustID)
	require.Equal(t, "onboarding-flow", decoded.Flow)

	wantDir := filepath.Join(repo, ".asagiri", "trust", result.TrustID)
	require.DirExists(t, wantDir)
	require.Equal(t, filepath.Join(wantDir, "report.md"), result.MDPath)
	require.Equal(t, filepath.Join(wantDir, "report.json"), result.JSONPath)

	mdBody, err := os.ReadFile(result.MDPath)
	require.NoError(t, err)
	md := string(mdBody)
	require.Contains(t, md, "# Trust Report")
	require.Contains(t, md, ConfidenceUnavailableLabel)
	require.NotContains(t, md, "Architecture: 0%")
	require.Contains(t, md, "## Uncovered zones")
	require.Contains(t, md, "architecture: no evidence (checks not run)")

	summary := FormatTerminalSummary(result.Report)
	require.Contains(t, summary, ConfidenceUnavailableLabel)
	require.Contains(t, summary, "Uncovered zones")
	require.Contains(t, summary, "architecture: no evidence (checks not run)")
	require.NotContains(t, summary, "Architecture:    0.00")
}

func TestStubAggregatorZeros(t *testing.T) {
	var a confidence.StubAggregator
	rep, err := a.Aggregate(context.Background(), nil)
	require.NoError(t, err)
	require.Zero(t, rep.Overall)
	require.NotEmpty(t, rep.Limits)
	require.Len(t, rep.UncoveredZones, 6)
}

func TestStubConfidenceAggregator(t *testing.T) {
	var a stubConfidenceAggregator
	rep, err := a.Aggregate(context.Background(), []VerificationCheck{})
	require.NoError(t, err)
	require.Len(t, rep.UncoveredZones, 6)
}

var (
	_ TrustEngine             = (*Engine)(nil)
	_ VerificationCheckRunner = (*verificationCheckRunnerFunc)(nil)
	_ ConfidenceAggregator    = stubConfidenceAggregator{}
)

type verificationCheckRunnerFunc struct{}

func (verificationCheckRunnerFunc) Run(_ context.Context, _ VerificationScope) (VerificationCheck, error) {
	return VerificationCheck{}, nil
}
