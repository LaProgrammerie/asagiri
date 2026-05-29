package trust

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
)

func TestRunStrictTrustFailsOnBlockedGate(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri", "products", "demo", "flows"), 0o755))
	eng := NewEngineWithChecks(repo, checks.NewRegistry())
	eng.Gates = NewGateEvaluator(&config.VerificationConfig{
		Gates: map[string]config.GateProfile{
			"production": {MinConfidence: map[string]float64{"overall": 1.0}},
		},
	})
	_, err := RunStrictTrust(context.Background(), eng, "flow", "", "demo")
	require.Error(t, err)
	var ste *StrictTrustError
	require.ErrorAs(t, err, &ste)
}

func TestRunStrictTrustFailsWithoutGatesLowConfidence(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri", "products", "demo", "flows"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, ".asagiri", "products", "demo", "flows", "onboarding.flow.yaml"),
		[]byte("id: onboarding\nentry_screen: s1\nsteps:\n  - id: s1\n    screen: s1\n    action: go\n"),
		0o644,
	))
	eng := NewEngineWithChecks(repo, checks.NewRegistry())
	_, err := RunStrictTrust(context.Background(), eng, "onboarding", "", "demo")
	require.Error(t, err)
	var ste *StrictTrustError
	require.ErrorAs(t, err, &ste)
	require.Contains(t, ste.Reason, "below strict floor")
}

func TestRunStrictTrustFailsOnZeroConfidence(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri", "products", "demo", "flows"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(repo, ".asagiri", "products", "demo", "flows", "flow.flow.yaml"),
		[]byte("id: flow\nentry_screen: s1\nsteps:\n  - id: s1\n    screen: s1\n    action: go\n"),
		0o644,
	))
	eng := NewEngineWithChecks(repo, checks.NewRegistry())
	_, err := RunStrictTrust(context.Background(), eng, "flow", "", "demo")
	require.Error(t, err)
	var ste *StrictTrustError
	require.ErrorAs(t, err, &ste)
	require.Contains(t, ste.Reason, "0%")
}
