package knowledge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestGoldenOnboardingFlowBuild(t *testing.T) {
	fixtureRoot := filepath.Join("testdata", "knowledge-graph", "onboarding-flow", "fixture")
	repo := t.TempDir()
	copyDir(t, fixtureRoot, repo)

	result, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:         repo,
		IncludeFlows:     true,
		IncludeContracts: true,
		IncludeCode:      true,
		IncludeTests:     true,
	})
	require.NoError(t, err)
	require.Greater(t, result.Nodes, 0)

	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	wantNodes := []string{
		"flow:onboarding",
		"action:invite_member",
		"api_operation:POST_invitations",
		"event:member.invited",
		"symbol:InvitationService_invite",
		"test:InvitationServiceTest",
	}
	for _, id := range wantNodes {
		_, err := store.GetNode(context.Background(), id)
		require.NoError(t, err, "missing node %s", id)
	}

	wantEdges := []struct {
		from, to string
		typ      knowledge.EdgeType
	}{
		{"flow:onboarding", "action:invite_member", knowledge.EdgeTypeRequires},
		{"action:invite_member", "api_operation:POST_invitations", knowledge.EdgeTypeRequires},
	}
	ctx := context.Background()
	for _, we := range wantEdges {
		edgeID := knowledge.EdgeID(we.typ, we.from, we.to)
		edge, err := store.GetEdge(ctx, edgeID)
		require.NoError(t, err, "missing edge %s", edgeID)
		require.Equal(t, we.from, edge.From)
		require.Equal(t, we.to, edge.To)
		require.Equal(t, we.typ, edge.Type)
	}

	q := knowledge.NewQuerier(store)
	impl, err := q.QueryImplements(ctx, "invite_member")
	require.NoError(t, err)
	implIDs := make(map[string]struct{})
	for _, n := range impl.Nodes {
		implIDs[n.ID] = struct{}{}
	}
	require.Contains(t, implIDs, "action:invite_member")
	require.Contains(t, implIDs, "api_operation:POST_invitations")

	_, err = os.Stat(filepath.Join(repo, ".asagiri", "knowledge", "graph.sqlite"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(repo, ".asagiri", "knowledge", "graph.json"))
	require.NoError(t, err)
}

func TestBuilderRepoRootWithCodeAndTests(t *testing.T) {
	repoRoot := findRepoRoot(t)
	if repoRoot == "" {
		t.Skip("repo root not found")
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "application", "internal", "config", "config.go")); err != nil {
		t.Skip("application sources not present")
	}
	if _, err := os.Stat(filepath.Join(repoRoot, ".asagiri", "products", "workspace-saas")); err != nil {
		t.Skip("workspace-saas product not present")
	}

	repo := t.TempDir()
	copyDir(t, filepath.Join(repoRoot, ".asagiri", "products", "workspace-saas"),
		filepath.Join(repo, ".asagiri", "products", "workspace-saas"))
	copyDir(t, filepath.Join(repoRoot, "application"), filepath.Join(repo, "application"))

	result, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:         repo,
		Scope:            "workspace-saas",
		IncludeFlows:     true,
		IncludeContracts: true,
		IncludeCode:      true,
		IncludeTests:     true,
	})
	require.NoError(t, err)
	require.Greater(t, result.Nodes, 0)
	require.Greater(t, result.Edges, 0)
	require.Less(t, result.Edges, 50000, "test linker should not explode edge count")
}

func TestBuilderIntegrationWorkspaceSaaS(t *testing.T) {
	repoRoot := findRepoRoot(t)
	if repoRoot == "" {
		t.Skip("repo root not found")
	}
	productDir := filepath.Join(repoRoot, ".asagiri", "products", "workspace-saas")
	if _, err := os.Stat(productDir); err != nil {
		t.Skip("workspace-saas product not present")
	}

	repo := t.TempDir()
	copyDir(t, filepath.Join(repoRoot, ".asagiri", "products", "workspace-saas"),
		filepath.Join(repo, ".asagiri", "products", "workspace-saas"))

	result, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:         repo,
		Scope:            "workspace-saas",
		IncludeFlows:     true,
		IncludeContracts: true,
	})
	require.NoError(t, err)
	require.Greater(t, result.Edges, 0)
}

func TestParseQueryPhraseImplements(t *testing.T) {
	parsed, ok := knowledge.ParseQueryPhrase("what implements invite_member?")
	require.True(t, ok)
	require.Equal(t, "implements:invite_member", parsed.Label)
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".asagiri", "products")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dst, 0o755))
	entries, err := os.ReadDir(src)
	require.NoError(t, err)
	for _, ent := range entries {
		srcPath := filepath.Join(src, ent.Name())
		dstPath := filepath.Join(dst, ent.Name())
		if ent.IsDir() {
			copyDir(t, srcPath, dstPath)
			continue
		}
		body, err := os.ReadFile(srcPath)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dstPath, body, 0o644))
	}
}
