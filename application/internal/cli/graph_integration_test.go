package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func copyGraphFixtureProduct(t *testing.T, repo string) {
	t.Helper()
	src := filepath.Join("..", "trust", "checks", "testdata", "minimal-product")
	dest := filepath.Join(repo, ".asagiri", "products", "minimal-product")
	require.NoError(t, copyDirGraph(src, dest))
	graphsSrc := filepath.Join("..", "trust", "checks", "testdata", "graphs-minimal.json")
	analysisDir := filepath.Join(repo, ".asagiri", "analysis", "minimal-product")
	require.NoError(t, os.MkdirAll(analysisDir, 0o755))
	require.NoError(t, copyFileGraph(graphsSrc, filepath.Join(analysisDir, "graphs.json")))
}

func copyFileGraph(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyDirGraph(src, dst string) error {
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
		return copyFileGraph(path, target)
	})
}

func graphConfigYAML() string {
	return `project:
  name: graph-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
verification:
  default_profile: production
  gates:
    production:
      min_confidence:
        overall: 0.0
      required_checks: []
execution_graph:
  enabled: true
  max_parallel: 2
  stop_on_risk: critical
  gates:
    human_approval_for: []
`
}

func TestCLIIntegrationGraphCommands(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), graphConfigYAML())
	copyGraphFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{"plan", "graph", "minimal-product", "--flow", "workspace-onboarding"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Asagiri Execution Graph")
	require.Contains(t, output.String(), "minimal-product")

	var graphID string
	entries, err := os.ReadDir(filepath.Join(repo, ".asagiri", "graphs"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	graphID = entries[0].Name()

	graphDir := filepath.Join(repo, ".asagiri", "graphs", graphID)
	require.FileExists(t, filepath.Join(graphDir, "execution-graph.yaml"))
	require.FileExists(t, filepath.Join(graphDir, "execution-graph.json"))
	require.FileExists(t, filepath.Join(graphDir, "plan.md"))
	require.FileExists(t, filepath.Join(graphDir, "metrics.json"))
	require.FileExists(t, filepath.Join(graphDir, "timeline.jsonl"))
	require.FileExists(t, filepath.Join(graphDir, "events.jsonl"))

	output.Reset()
	root.SetArgs([]string{"plan", "explain", "minimal-product", "--flow", "workspace-onboarding"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Execution Plan Explanation")

	for _, tc := range []struct {
		format   string
		contains string
	}{
		{"mermaid", "graph TD"},
		{"json", `"product": "minimal-product"`},
		{"dot", "digraph execution_graph"},
		{"markdown", "# Execution Graph"},
	} {
		t.Run("visualize_"+tc.format, func(t *testing.T) {
			output.Reset()
			root.SetArgs([]string{"graph", "visualize", graphID, "--format", tc.format})
			require.NoError(t, root.Execute(), output.String())
			require.Contains(t, output.String(), tc.contains)
		})
	}

	output.Reset()
	root.SetArgs([]string{"plan", "graph", "minimal-product", "--flow", "workspace-onboarding", "--output", "markdown"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "# Execution Graph")

	output.Reset()
	root.SetArgs([]string{"graph", "run", "minimal-product", "--flow", "workspace-onboarding", "--dry-run"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Dry-run")

	jsonOut := new(bytes.Buffer)
	root.SetOut(jsonOut)
	root.SetErr(jsonOut)
	root.SetArgs([]string{"plan", "graph", "minimal-product", "--flow", "workspace-onboarding", "--ci", "--json"})
	require.NoError(t, root.Execute(), jsonOut.String())
	var planResult PlanGraphResult
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &planResult))
	require.Equal(t, "minimal-product", planResult.Graph.Product)
	require.NotEmpty(t, planResult.Schedule.ParallelGroups)

	jsonOut.Reset()
	root.SetArgs([]string{"graph", "run", "minimal-product", "--flow", "workspace-onboarding", "--dry-run", "--ci", "--json"})
	require.NoError(t, root.Execute(), jsonOut.String())
	var runResult GraphRunJSONResult
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &runResult))
	require.True(t, runResult.DryRun)
	require.Equal(t, executiongraph.GraphStatusReady, runResult.Result.Status)

	writePausedGraphWithCheckpoint(t, repo, graphID)

	jsonOut.Reset()
	root.SetArgs([]string{"graph", "resume", graphID, "--json"})
	require.NoError(t, root.Execute(), jsonOut.String())
	var resumeResult GraphResumeResult
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &resumeResult))
	require.Equal(t, executiongraph.GraphStatusCompleted, resumeResult.Result.Status)

	root.SetOut(output)
	root.SetErr(output)
	output.Reset()
	root.SetArgs([]string{"graph", "status", graphID})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Status: completed")

	jsonOut.Reset()
	root.SetOut(jsonOut)
	root.SetErr(jsonOut)
	root.SetArgs([]string{"graph", "status", graphID, "--json"})
	require.NoError(t, root.Execute(), jsonOut.String())
	var statusResult GraphStatusResult
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &statusResult))
	require.Equal(t, executiongraph.GraphStatusCompleted, statusResult.Graph.Status)
	require.Equal(t, "minimal-product", statusResult.Graph.Product)
}

func TestCLIIntegrationGraphFlowRequired(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), graphConfigYAML())
	copyGraphFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	for _, args := range [][]string{
		{"plan", "graph", "minimal-product"},
		{"plan", "explain", "minimal-product"},
		{"graph", "run", "minimal-product", "--dry-run"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			output.Reset()
			root.SetArgs(args)
			err := root.Execute()
			require.Error(t, err)
			require.ErrorIs(t, err, errGraphFlowRequired)
		})
	}
}

func TestCLIIntegrationGraphDisabled(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), `project:
  name: graph-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
execution_graph:
  enabled: false
`)
	copyGraphFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	for _, args := range [][]string{
		{"plan", "graph", "minimal-product", "--flow", "workspace-onboarding"},
		{"graph", "run", "minimal-product", "--flow", "workspace-onboarding", "--dry-run"},
		{"graph", "status", "graph-00000000-00000000"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			output.Reset()
			root.SetArgs(args)
			err := root.Execute()
			require.Error(t, err)
			require.ErrorIs(t, err, errGraphNotEnabled)
		})
	}
}

func TestCLIIntegrationGraphResumeHonorsPersistedStrategy(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), graphConfigYAML())
	copyGraphFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"plan", "graph", "minimal-product", "--flow", "workspace-onboarding"})
	require.NoError(t, root.Execute(), output.String())

	entries, err := os.ReadDir(filepath.Join(repo, ".asagiri", "graphs"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	graphID := entries[0].Name()

	setGraphRunStrategy(t, repo, graphID, true, executiongraph.CheckpointEveryGroup)
	writePausedGraphWithCheckpoint(t, repo, graphID)

	repoObj := executiongraph.NewRepository(repo)
	loaded, err := repoObj.Load(graphID)
	require.NoError(t, err)
	require.True(t, loaded.Strategy.StrictTrust)
	require.Equal(t, executiongraph.CheckpointEveryGroup, loaded.Strategy.CheckpointEvery)
	opts := graphRunOptionsFromPersisted(loaded)
	require.True(t, opts.StrictTrust)
	require.Equal(t, executiongraph.CheckpointEveryGroup, opts.CheckpointEvery)

	jsonOut := new(bytes.Buffer)
	root.SetOut(jsonOut)
	root.SetErr(jsonOut)
	root.SetArgs([]string{"graph", "resume", graphID, "--json"})
	require.NoError(t, root.Execute(), jsonOut.String())
	var resumeResult GraphResumeResult
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &resumeResult))
	require.Equal(t, executiongraph.GraphStatusCompleted, resumeResult.Result.Status)
}

func TestCLIIntegrationGraphCheckpointEveryNode(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), graphConfigYAML())
	copyGraphFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"graph", "run", "minimal-product", "--flow", "workspace-onboarding", "--checkpoint-every", "node"})
	require.NoError(t, root.Execute(), output.String())

	entries, err := os.ReadDir(filepath.Join(repo, ".asagiri", "graphs"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	graphID := entries[0].Name()

	repoObj := executiongraph.NewRepository(repo)
	loaded, err := repoObj.Load(graphID)
	require.NoError(t, err)
	executed := 0
	for _, n := range loaded.Nodes {
		if n.Status == executiongraph.NodeStatusSucceeded {
			executed++
		}
	}
	require.Greater(t, executed, 0)
	count, err := repoObj.CountCheckpoints(graphID)
	require.NoError(t, err)
	require.Equal(t, executed, count)
}

func TestCLIIntegrationGraphCheckpointEveryGroup(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), graphConfigYAML())
	copyGraphFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"graph", "run", "minimal-product", "--flow", "workspace-onboarding", "--checkpoint-every", "group"})
	require.NoError(t, root.Execute(), output.String())

	entries, err := os.ReadDir(filepath.Join(repo, ".asagiri", "graphs"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	graphID := entries[0].Name()

	repoObj := executiongraph.NewRepository(repo)
	loaded, err := repoObj.Load(graphID)
	require.NoError(t, err)
	require.Equal(t, executiongraph.CheckpointEveryGroup, loaded.Strategy.CheckpointEvery)

	executed := 0
	for _, n := range loaded.Nodes {
		if n.Status == executiongraph.NodeStatusSucceeded {
			executed++
		}
	}
	require.Greater(t, executed, 0)

	sched, err := executiongraph.DefaultScheduler{}.Schedule(t.Context(), executiongraph.ScheduleRequest{Graph: loaded})
	require.NoError(t, err)
	count, err := repoObj.CountCheckpoints(graphID)
	require.NoError(t, err)
	require.Less(t, count, executed)
	require.Equal(t, len(sched.ParallelGroups), count)
}

func writePausedGraphWithCheckpoint(t *testing.T, repo, graphID string) {
	t.Helper()
	writePausedGraphFixture(t, repo, graphID)
	path := filepath.Join(repo, ".asagiri", "graphs", graphID, "execution-graph.yaml")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var graph executiongraph.ExecutionGraph
	require.NoError(t, yaml.Unmarshal(raw, &graph))
	afterNode := graph.Nodes[0].ID
	for _, n := range graph.Nodes {
		if n.Type == executiongraph.NodeTypeImplementation && n.Status != executiongraph.NodeStatusSucceeded {
			afterNode = n.ID
			break
		}
	}
	for _, n := range graph.Nodes {
		if n.Type == executiongraph.NodeTypeInvestigation {
			afterNode = n.ID
		}
	}
	repoObj := executiongraph.NewRepository(repo)
	_, err = repoObj.SaveCheckpoint(graphID, executiongraph.CheckpointState{
		AfterNode: afterNode,
		CreatedAt: "2026-05-29T12:00:00Z",
	})
	require.NoError(t, err)
}

func setGraphRunStrategy(t *testing.T, repo, graphID string, strictTrust bool, checkpointEvery string) {
	t.Helper()
	path := filepath.Join(repo, ".asagiri", "graphs", graphID, "execution-graph.yaml")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var graph executiongraph.ExecutionGraph
	require.NoError(t, yaml.Unmarshal(raw, &graph))
	graph.Strategy.StrictTrust = strictTrust
	graph.Strategy.CheckpointEvery = checkpointEvery
	body, err := yaml.Marshal(&graph)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, body, 0o644))
	jsonPath := filepath.Join(repo, ".asagiri", "graphs", graphID, "execution-graph.json")
	jsonBody, err := json.MarshalIndent(graph, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(jsonPath, jsonBody, 0o644))
}

func writePausedGraphFixture(t *testing.T, repo, graphID string) {
	t.Helper()
	path := filepath.Join(repo, ".asagiri", "graphs", graphID, "execution-graph.yaml")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var graph executiongraph.ExecutionGraph
	require.NoError(t, yaml.Unmarshal(raw, &graph))
	graph.Status = executiongraph.GraphStatusPaused
	for i := range graph.Nodes {
		if graph.Nodes[i].Type == executiongraph.NodeTypeTrustVerification {
			graph.Nodes[i].Status = executiongraph.NodeStatusSucceeded
		}
	}
	body, err := yaml.Marshal(&graph)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, body, 0o644))
	jsonPath := filepath.Join(repo, ".asagiri", "graphs", graphID, "execution-graph.json")
	jsonBody, err := json.MarshalIndent(graph, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(jsonPath, jsonBody, 0o644))
}

func TestCLIIntegrationGraphCIFailsOnBudget(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), graphConfigYAML())
	copyGraphFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"graph", "run", "minimal-product", "--flow", "workspace-onboarding", "--dry-run", "--ci", "--budget", "0.01"})
	err = root.Execute()
	require.Error(t, err)
	require.ErrorIs(t, err, errGraphCIFailed)
}
