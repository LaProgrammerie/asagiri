package doctor_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/doctor"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestBuildArchitectureCrossGapsFixture(t *testing.T) {
	repo := setupArchitectureFixture(t)

	report, err := doctor.BuildArchitecture(repo)
	require.NoError(t, err)
	require.Equal(t, doctor.ArchitectureReportVersion, report.ReportVersion)
	requirePathsEqual(t, repo, report.Repository.GitRoot)

	require.Equal(t, 2, report.Summary.Tasks)
	require.Equal(t, 1, report.Summary.ExecutionGraphs)
	require.Equal(t, 3, report.Summary.ExecutionGraphNodes)
	require.Equal(t, 1, report.Summary.TasksWithoutGraphNode)
	require.Equal(t, 3, report.Summary.GraphNodesNeverExecuted)
	require.Equal(t, 2, report.Summary.TasksWithoutKnowledge)
	require.Equal(t, 1, report.Summary.AgentRunsWithoutTask)
	require.Equal(t, 1, report.Summary.TrustGapsCriticalFlows)

	kinds := map[string]int{}
	for _, f := range report.Findings {
		kinds[f.Kind]++
	}
	require.Equal(t, 1, kinds["task_without_graph_node"])
	require.Equal(t, 3, kinds["graph_node_never_executed"])
	require.Equal(t, 2, kinds["task_without_knowledge_context"])
	require.Equal(t, 1, kinds["agent_run_without_task"])
	require.Equal(t, 1, kinds["trust_gap_critical_flow"])

	require.NotEmpty(t, report.Recommendations)
	clis := recommendationCLIs(report.Recommendations)
	require.Contains(t, clis, "asa knowledge build")
	require.Contains(t, clis, "asa agents runs --json")
	require.True(t, containsPrefix(clis, "asa plan graph"))
	require.True(t, containsPrefix(clis, "asa verify trust"))

	stable := normalizeArchitectureReportForGolden(report)
	raw1, err := json.Marshal(stable)
	require.NoError(t, err)
	raw2, err := json.Marshal(stable)
	require.NoError(t, err)
	require.Equal(t, string(raw1), string(raw2))

	golden := filepath.Join("testdata", "architecture", "expected.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		enc := json.NewEncoder(mustCreate(t, golden))
		enc.SetIndent("", "  ")
		require.NoError(t, enc.Encode(stable))
		return
	}
	want, err := os.ReadFile(golden)
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(raw1))
}

func TestBuildArchitectureEmptyRepo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	report, err := doctor.BuildArchitecture(dir)
	require.NoError(t, err)
	require.Equal(t, doctor.ArchitectureReportVersion, report.ReportVersion)
	require.Empty(t, report.Findings)
}

func setupArchitectureFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	initGitRepo(t, dir)

	cfgDir := filepath.Join(dir, ".asagiri")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`project:
  name: arch-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
worktrees:
  base_path: .asagiri/worktrees
`), 0o644))

	dbPath := filepath.Join(cfgDir, "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.Migrate())
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-arch", Feature: "feat-arch", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-001", RunID: "run-arch", Feature: "feat-arch", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-orphan", RunID: "run-arch", Feature: "feat-arch", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	require.NoError(t, store.Close())

	graphSrc := filepath.Join("testdata", "architecture", "execution-graph.json")
	graphData, err := os.ReadFile(graphSrc)
	require.NoError(t, err)
	graphDir := filepath.Join(cfgDir, "graphs", "graph-2026-05-27-a1b2c3d4")
	require.NoError(t, os.MkdirAll(graphDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(graphDir, "execution-graph.json"), graphData, 0o644))

	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-001", RunID: "run-arch", Feature: "feat-arch", AgentID: "dev", ExitCode: 0,
	}))
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		RunID: "run-arch", Feature: "feat-arch", AgentID: "dev", ExitCode: 0,
	}))

	return dir
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func recommendationCLIs(actions []doctor.Action) []string {
	out := make([]string, 0, len(actions))
	for _, a := range actions {
		out = append(out, a.CLI)
	}
	return out
}

func containsPrefix(clis []string, prefix string) bool {
	for _, c := range clis {
		if len(c) >= len(prefix) && c[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func normalizeArchitectureReportForGolden(report doctor.ArchitectureReport) doctor.ArchitectureReport {
	if report.Repository.GitRoot != "" {
		report.Repository.GitRoot = "<repo>"
	}
	return report
}

func requirePathsEqual(t *testing.T, want, got string) {
	t.Helper()
	wantEval, err := filepath.EvalSymlinks(want)
	require.NoError(t, err)
	gotEval, err := filepath.EvalSymlinks(got)
	require.NoError(t, err)
	require.Equal(t, wantEval, gotEval)
}

func mustCreate(t *testing.T, path string) *os.File {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	f, err := os.Create(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}
