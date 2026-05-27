package product

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAndValidateYAML(t *testing.T) {
	p, err := ParseProductYAML([]byte("name: demo\nintent: build onboarding\nstack: react\n"))
	require.NoError(t, err)
	require.Equal(t, "demo", p.Name)

	_, err = ParseFlowYAML([]byte("id: f1\nentry_screen: s1\nsteps:\n  - id: x\n    screen: s1\n    action: submit\n    sensitive: true\n"))
	require.Error(t, err)

	s, err := ParseScreenYAML([]byte("id: s1\ntitle: Home\nroute: /\n"))
	require.NoError(t, err)
	require.Equal(t, "/", s.Route)

	_, err = ParseBusinessYAML([]byte("objective:\n  primary: reduce onboarding friction\nsuccess_metrics:\n  - id: onboarding_completion_rate\n    target: \">=70%\"\n"))
	require.NoError(t, err)
}

func TestRepositoryPathTraversalGuard(t *testing.T) {
	repo := NewRepository(t.TempDir())
	err := repo.writeSafe(filepath.Join(t.TempDir(), "base"), "../bad.txt", []byte("x"))
	require.Error(t, err)
}

func TestProductLayerGoldenAndGeneration(t *testing.T) {
	repoRoot := t.TempDir()
	svc := NewService(repoRoot)

	productName, err := svc.CreatePrototype(CreatePrototypeOptions{
		Intent:  "workspace onboarding",
		Product: "simple-onboarding",
	})
	require.NoError(t, err)
	require.Equal(t, "simple-onboarding", productName)

	require.NoError(t, svc.ExtractFlows(productName, false))
	summary, err := svc.InspectFlows(productName)
	require.NoError(t, err)
	require.NoError(t, svc.ExtractContracts(productName, false))
	_, err = svc.ReviewFlows(productName, false)
	require.NoError(t, err)
	_, err = svc.DeriveArchitecture(productName, false)
	require.NoError(t, err)
	require.NoError(t, svc.GenerateSpecFromProduct(productName, false))
	_, err = svc.ReviewProduct(productName, false)
	require.NoError(t, err)

	flowPath := filepath.Join(repoRoot, ".asagiri", "products", productName, "flows", "workspace-onboarding.flow.yaml")
	require.FileExists(t, flowPath)
	flowBody, err := os.ReadFile(flowPath)
	require.NoError(t, err)
	require.Contains(t, string(flowBody), readGolden(t, "expected-flow-id.txt"))

	require.Equal(t, readGolden(t, "expected-inspect.txt"), strings.TrimSpace(summary))
	require.FileExists(t, filepath.Join(repoRoot, ".asagiri", "specs", productName, "spec.md"))
	require.FileExists(t, filepath.Join(repoRoot, ".asagiri", "tasks", productName, productName+"-001.yaml"))
	require.FileExists(t, filepath.Join(repoRoot, ".kiro", "specs", productName, "tasks.md"))
	require.FileExists(t, filepath.Join(repoRoot, ".asagiri", "products", productName, "business.yaml"))
	require.FileExists(t, filepath.Join(repoRoot, ".asagiri", "products", productName, "reviews", "flow-review.md"))
	require.FileExists(t, filepath.Join(repoRoot, ".asagiri", "products", productName, "reviews", "architecture-review.md"))
}

func TestExtractContractsHardFailsOnUncoupledMetrics(t *testing.T) {
	repoRoot := t.TempDir()
	svc := NewService(repoRoot)
	productName, err := svc.CreatePrototype(CreatePrototypeOptions{
		Intent:  "workspace onboarding",
		Product: "hard-fail-metrics",
	})
	require.NoError(t, err)

	modelPath := filepath.Join(repoRoot, ".asagiri", "products", productName, "extraction", "extracted-model.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(modelPath), 0o755))
	raw := []byte(`product: hard-fail-metrics
business:
  objective:
    primary: reduce onboarding friction
flows:
  - id: onboarding
    title: Onboarding
    entry_screen: landing
    steps:
      - id: step-1
        screen: landing
        action: invite_member
        contract_ref: POST /api/invite
screens:
  - id: landing
    title: Landing
    route: /
`)
	require.NoError(t, os.WriteFile(modelPath, raw, 0o644))

	err = svc.ExtractContracts(productName, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "strict coupling failed")
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", "product-layer", "simple-onboarding", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return strings.TrimSpace(string(data))
}
