package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
)

func knowledgeConfigYAML() string {
	return `project:
  name: knowledge-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
`
}

func copyKnowledgeOnboardingFixture(t *testing.T, repo string) {
	t.Helper()
	src := filepath.Join("..", "knowledge", "testdata", "knowledge-graph", "onboarding-flow", "fixture")
	dest := repo
	require.NoError(t, copyDirGraph(src, dest))
}

func TestCLIIntegrationKnowledgeBuildAndQuery(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), knowledgeConfigYAML())
	copyKnowledgeOnboardingFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{
		"knowledge", "build",
		"--include-flows",
		"--include-contracts",
		"--scope", "workspace-saas",
	})
	require.NoError(t, root.Execute(), output.String())
	out := output.String()
	require.Contains(t, out, "Asagiri Knowledge Graph")
	require.Contains(t, out, "Nodes:")
	require.Contains(t, out, "Sources:")
	_, err = os.Stat(filepath.Join(repo, ".asagiri", "knowledge", "graph.sqlite"))
	require.NoError(t, err)

	output.Reset()
	root.SetArgs([]string{"knowledge", "query", "what implements invite_member?"})
	require.NoError(t, root.Execute(), output.String())
	queryOut := output.String()
	require.Contains(t, queryOut, "action:invite_member")
	require.Contains(t, queryOut, "api_operation:POST_invitations")
}

func TestCLIIntegrationKnowledgeBuildJSON(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), knowledgeConfigYAML())
	copyKnowledgeOnboardingFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{
		"knowledge", "build",
		"--include-flows",
		"--include-contracts",
		"--scope", "workspace-saas",
		"--json",
	})
	require.NoError(t, root.Execute(), output.String())

	var result knowledge.BuildResult
	require.NoError(t, json.Unmarshal(output.Bytes(), &result))
	require.Greater(t, result.Nodes, 0)
	require.Greater(t, result.Edges, 0)
}

func loadKnowledgeGraphFixture(t *testing.T, repo string) {
	t.Helper()
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	body, err := os.ReadFile(filepath.Join("..", "knowledge", "testdata", "knowledge-graph", "onboarding-flow", "graph.json"))
	require.NoError(t, err)
	graph, err := knowledge.ParseJSON(body)
	require.NoError(t, err)

	ctx := context.Background()
	for _, node := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, node))
	}
	for _, edge := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, edge))
	}
}

func TestCLIIntegrationImpactAnalyze(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), knowledgeConfigYAML())
	loadKnowledgeGraphFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{"impact", "analyze", "--flow", "onboarding", "--action", "invite_member"})
	require.NoError(t, root.Execute(), output.String())
	out := output.String()
	require.Contains(t, out, "Impact Analysis")
	require.Contains(t, out, "onboarding / invite_member")
	require.Contains(t, out, "POST /invitations")
	require.Contains(t, out, "member.invited")
	require.Contains(t, out, "InvitationServiceTest")
}

func TestCLIIntegrationKnowledgeExplain(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), knowledgeConfigYAML())
	loadKnowledgeGraphFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{"knowledge", "explain", "onboarding", "invite_member", "InvitationService"})
	require.NoError(t, root.Execute(), output.String())
	out := output.String()
	require.Contains(t, out, "Knowledge path")
	require.Contains(t, out, "invite_member")
	require.Contains(t, out, "InvitationService")
}

func TestCLIIntegrationKnowledgeQueryRejectsUnknownPhrase(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), knowledgeConfigYAML())
	copyKnowledgeOnboardingFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{
		"knowledge", "build",
		"--include-flows",
		"--include-contracts",
		"--scope", "workspace-saas",
	})
	require.NoError(t, root.Execute())

	output.Reset()
	root.SetArgs([]string{"knowledge", "query", "show me everything"})
	err = root.Execute()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "phrase non reconnue"))
}

func TestCLIIntegrationKnowledgeSnapshot(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), knowledgeConfigYAML())
	copyKnowledgeOnboardingFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{
		"knowledge", "build",
		"--include-flows",
		"--include-contracts",
		"--scope", "workspace-saas",
	})
	require.NoError(t, root.Execute())

	output.Reset()
	root.SetArgs([]string{"knowledge", "snapshot", "--name", "smoke", "--json"})
	require.NoError(t, root.Execute(), output.String())

	var snap knowledge.SnapshotResult
	require.NoError(t, json.Unmarshal(output.Bytes(), &snap))
	require.Equal(t, "smoke", snap.Name)
	_, err = os.Stat(filepath.Join(repo, ".asagiri", "knowledge", "snapshots", "smoke", "graph.json"))
	require.NoError(t, err)
}
