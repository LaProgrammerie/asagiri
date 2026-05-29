package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	"github.com/stretchr/testify/require"
)

func TestFlowExtractorMissingFlowsDir(t *testing.T) {
	repo := t.TempDir()
	nodes, edges, warnings, err := extractors.FlowExtractor{}.Extract(context.Background(), repo, "demo")
	require.NoError(t, err)
	require.Empty(t, nodes)
	require.Empty(t, edges)
	require.Empty(t, warnings)
}

func TestFlowExtractorRejectsMissingFlowID(t *testing.T) {
	repo := t.TempDir()
	flowDir := filepath.Join(repo, ".asagiri", "products", "demo", "flows")
	require.NoError(t, os.MkdirAll(flowDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "bad.flow.yaml"), []byte("steps:\n  - id: s1\n"), 0o644))

	_, _, _, err := extractors.FlowExtractor{}.Extract(context.Background(), repo, "demo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "flow id required")
}

func TestFlowExtractorWarnsOnStepWithoutID(t *testing.T) {
	repo := t.TempDir()
	flowDir := filepath.Join(repo, ".asagiri", "products", "demo", "flows")
	require.NoError(t, os.MkdirAll(flowDir, 0o755))
	body := []byte("id: onboarding\nsteps:\n  - action: noop\n")
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "onboarding.flow.yaml"), body, 0o644))

	_, _, warnings, err := extractors.FlowExtractor{}.Extract(context.Background(), repo, "demo")
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	require.Contains(t, warnings[0], "step without id skipped")
}

func TestFlowExtractorWarnsOnUnresolvedContractRef(t *testing.T) {
	repo := t.TempDir()
	flowDir := filepath.Join(repo, ".asagiri", "products", "demo", "flows")
	require.NoError(t, os.MkdirAll(flowDir, 0o755))
	body := []byte("id: onboarding\nsteps:\n  - id: s1\n    action: act\n    contract_ref: not-a-method\n")
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "onboarding.flow.yaml"), body, 0o644))

	_, _, warnings, err := extractors.FlowExtractor{}.Extract(context.Background(), repo, "demo")
	require.NoError(t, err)
	require.NotEmpty(t, warnings)
	require.True(t, strings.Contains(warnings[0], "unresolved contract_ref"))
}

func TestContractExtractorMissingDir(t *testing.T) {
	repo := t.TempDir()
	nodes, edges, warnings, err := extractors.ContractExtractor{}.Extract(context.Background(), repo, "demo")
	require.NoError(t, err)
	require.Empty(t, nodes)
	require.Empty(t, edges)
	require.Empty(t, warnings)
}

func TestContractExtractorSkipsInvalidOpenAPIWithWarning(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, ".asagiri", "products", "demo", "contracts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.openapi.yaml"), []byte("openapi: 3.1.0\n"), 0o644))

	nodes, _, warnings, err := extractors.ContractExtractor{}.Extract(context.Background(), repo, "demo")
	require.NoError(t, err)
	require.Empty(t, nodes)
	require.NotEmpty(t, warnings)
	require.Contains(t, warnings[0], "contracts: skip")
}

func TestEventExtractorMissingFile(t *testing.T) {
	repo := t.TempDir()
	nodes, edges, warnings, err := extractors.EventExtractor{}.Extract(context.Background(), repo, "demo")
	require.NoError(t, err)
	require.Empty(t, nodes)
	require.Empty(t, edges)
	require.Empty(t, warnings)
}

func TestEventExtractorInvalidYAML(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, ".asagiri", "products", "demo", "contracts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "events.yaml"), []byte("events: [\n"), 0o644))

	_, _, _, err := extractors.EventExtractor{}.Extract(context.Background(), repo, "demo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "events:")
}

func TestEventExtractorWarnsWhenNoEventsSection(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, ".asagiri", "products", "demo", "contracts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "events.yaml"), []byte("version: 1\n"), 0o644))

	nodes, _, warnings, err := extractors.EventExtractor{}.Extract(context.Background(), repo, "demo")
	require.NoError(t, err)
	require.Empty(t, nodes)
	require.Contains(t, warnings[0], "no events section found")
}

func TestPermissionExtractorInvalidYAML(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, ".asagiri", "products", "demo", "contracts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "permissions.yaml"), []byte("permissions: [\n"), 0o644))

	_, _, _, err := extractors.PermissionExtractor{}.Extract(context.Background(), repo, "demo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "permissions:")
}
