package product

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateE2ETestPlaywright(t *testing.T) {
	repoRoot := t.TempDir()
	svc := NewService(repoRoot)
	productName, err := svc.CreatePrototype(CreatePrototypeOptions{
		Intent:  "workspace onboarding",
		Product: "e2e-fixture",
	})
	require.NoError(t, err)
	require.NoError(t, svc.ExtractFlows(productName, false))

	result, err := svc.GenerateE2ETest(E2EGeneratorOptions{
		Product: productName,
		FlowID:  "workspace-onboarding",
		Runner:  "playwright",
	})
	require.NoError(t, err)
	require.Equal(t, "playwright", result.Runner)
	require.Greater(t, result.StepCount, 0)
	require.FileExists(t, result.Path)

	body, err := os.ReadFile(result.Path)
	require.NoError(t, err)
	content := string(body)
	require.Contains(t, content, "@playwright/test")
	require.Contains(t, content, "workspace-onboarding")
	require.Contains(t, content, "click_get_started")
}

func TestGenerateE2ETestCypressDryRun(t *testing.T) {
	repoRoot := t.TempDir()
	svc := NewService(repoRoot)
	productName, err := svc.CreatePrototype(CreatePrototypeOptions{
		Intent:  "workspace onboarding",
		Product: "e2e-cypress",
	})
	require.NoError(t, err)
	require.NoError(t, svc.ExtractFlows(productName, false))

	result, err := svc.GenerateE2ETest(E2EGeneratorOptions{
		Product: productName,
		FlowID:  "workspace-onboarding",
		Runner:  "cypress",
		DryRun:  true,
	})
	require.NoError(t, err)
	require.Equal(t, "cypress", result.Runner)
	require.False(t, strings.HasSuffix(result.Path, ".spec.ts"))
	_, err = os.Stat(result.Path)
	require.Error(t, err)

	flowPath := filepath.Join(repoRoot, ".asagiri", "products", productName, "flows", "workspace-onboarding.flow.yaml")
	require.FileExists(t, flowPath)
}

func TestGenerateE2ETestMissingFlow(t *testing.T) {
	svc := NewService(t.TempDir())
	_, err := svc.GenerateE2ETest(E2EGeneratorOptions{
		Product: "missing",
		FlowID:  "no-such-flow",
	})
	require.Error(t, err)
}
