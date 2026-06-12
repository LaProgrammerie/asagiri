package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentanalytics"
	"github.com/LaProgrammerie/asagiri/application/internal/agentexternal"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/agentslist"
	"github.com/LaProgrammerie/asagiri/application/internal/agentsync"
	"github.com/stretchr/testify/require"
)

func TestAgentsListJSONAfterInit(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "list", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentslist.Report
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentslist.ReportVersion, report.ReportVersion)
	require.NotEmpty(t, report.Agents)
	for _, entry := range report.Agents {
		require.NotEmpty(t, entry.ID)
		require.Len(t, entry.ContentHash, 64)
	}
}

func TestAgentsListTextOutput(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "list"})
	require.NoError(t, root.Execute())
	body := out.String()
	require.Contains(t, body, "Asagiri Agents")
	require.Contains(t, body, "dev")
	require.Contains(t, body, "hash=")
}

func TestAgentsShowJSON(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "show", "dev", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var entry agentslist.Entry
	require.NoError(t, json.Unmarshal(out.Bytes(), &entry))
	require.Equal(t, "dev", entry.ID)
	require.NotEmpty(t, entry.ProviderSupport)
}

func TestAgentsListSnapshotGolden(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "list", "--json"})
	require.NoError(t, root.Execute())

	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	goldenPath := filepath.Join(filepath.Dir(testFile), "testdata", "agents_list_embedded.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0o755))
		require.NoError(t, os.WriteFile(goldenPath, out.Bytes(), 0o644))
		return
	}
	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	require.JSONEq(t, string(want), out.String())
}

func TestAgentsSyncDryRunJSON(t *testing.T) {
	trustWorkTestRepo(t)

	var out, errOut bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"agents", "sync", "--json"})
	require.NoError(t, root.Execute())
	requireStdoutSingleJSON(t, out.Bytes())

	var report agentsync.Report
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentsync.ReportVersion, report.ReportVersion)
	require.Equal(t, "check", report.Mode)
	require.False(t, report.Wrote)
}

func TestAgentsSyncWriteCreatesRegistry(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)

	root := newRootCmd()
	root.SetArgs([]string{"agents", "sync", "--write"})
	require.NoError(t, root.Execute())

	registry := filepath.Join(dir, ".asagiri", "agents")
	entries, err := os.ReadDir(registry)
	require.NoError(t, err)
	require.NotEmpty(t, entries)
}

func TestAgentsSyncConflictExitNonZero(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	devPath := filepath.Join(dir, ".asagiri", "agents", "dev.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(devPath), 0o755))
	require.NoError(t, os.WriteFile(devPath, []byte(`id: dev
version: "9.9.9"
role: dev
system_prompt: custom
output_contract:
  format: asagiri-v1
`), 0o644))

	root := newRootCmd()
	root.SetArgs([]string{"agents", "sync", "--write"})
	require.Error(t, root.Execute())
}

func TestAgentsSyncForceOverwrite(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	devPath := filepath.Join(dir, ".asagiri", "agents", "dev.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(devPath), 0o755))
	require.NoError(t, os.WriteFile(devPath, []byte(`id: dev
version: "9.9.9"
role: dev
system_prompt: custom
output_contract:
  format: asagiri-v1
`), 0o644))

	root := newRootCmd()
	root.SetArgs([]string{"agents", "sync", "--write", "--force", "--agent", "dev"})
	require.NoError(t, root.Execute())

	data, err := os.ReadFile(devPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "1.0.0")
}

func TestAgentsStatsJSONEmptyLedger(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "stats", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentanalytics.Report
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentanalytics.ReportVersion, report.ReportVersion)
	require.Equal(t, 0, report.Global.TotalRuns)
}

func TestAgentsStatsFilterByAgent(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "t1", AgentID: "dev", Provider: "exec", ExitCode: 0, DurationMS: 100,
	}))
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "t2", AgentID: "reviewer", Provider: "codex", ExitCode: 0, DurationMS: 200,
	}))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "stats", "--agent", "dev", "--json"})
	require.NoError(t, root.Execute())

	var report agentanalytics.Report
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, 1, report.Global.TotalRuns)
	require.Equal(t, "dev", report.Filter.AgentID)
}

func TestAgentsRunJSONNotFound(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "run", "missing-run-id", "--json"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "introuvable")
}

func TestAgentsRunJSONPresent(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	valid := true
	logDir := ".asagiri/logs/task-x/agents/dev"
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-x", RunID: "run-cli", Feature: "feat", AgentID: "dev",
		ContractValid: &valid, LogDir: logDir, ExitCode: 0, DurationMS: 50,
	}))
	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "prompt.md"), []byte("prompt"), 0o644))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "run", "run-cli", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentledger.InspectReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentledger.InspectReportVersion, report.ReportVersion)
	require.Equal(t, "run-cli", report.RunID)
	require.True(t, report.Artifacts[0].Exists)
}

func TestAgentsRunPreviewJSONPresent(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	logDir := ".asagiri/logs/task-p/agents/dev"
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-p", RunID: "run-preview", AgentID: "dev", LogDir: logDir,
		PromptHash: agentledger.HashText("p"), OutputHash: agentledger.HashText("o"),
	}))
	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "context.json"), []byte(`{"ok":true}`), 0o644))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "run", "run-preview", "--preview", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentledger.ReplayPreviewReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentledger.ReplayPreviewReportVersion, report.ReportVersion)
	require.Contains(t, report.Artifacts[2].Content, "ok")
}

func TestAgentsRunPreviewIncludePrompt(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	logDir := ".asagiri/logs/task-q/agents/dev"
	prompt := "full prompt body"
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-q", RunID: "run-prompt-cli", AgentID: "dev", LogDir: logDir,
	}))
	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "prompt.md"), []byte(prompt), 0o644))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "run", "run-prompt-cli", "--preview", "--include-prompt", "--json"})
	require.NoError(t, root.Execute())

	var report agentledger.ReplayPreviewReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, prompt, report.Artifacts[0].Content)
}

func TestAgentsExportJSONNotFound(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "export", "missing-run", "--output", t.TempDir(), "--json"})
	err := root.Execute()
	require.Error(t, err)
}

func TestAgentsExportJSONWithoutPrompt(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	runID := "run-export-cli"
	logDir := ".asagiri/logs/task-e/agents/dev"
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-e", RunID: runID, AgentID: "dev", LogDir: logDir,
		PromptHash: agentledger.HashText("p"),
	}))
	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "prompt.md"), []byte("p"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "context.json"), []byte("{}"), 0o644))

	outDir := filepath.Join(dir, "export-out")
	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "export", runID, "--output", outDir, "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentledger.ExportReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentledger.ExportReportVersion, report.ReportVersion)
	require.FileExists(t, filepath.Join(outDir, "manifest.json"))
	require.FileExists(t, filepath.Join(outDir, "artifacts", "prompt.md"))
}

func TestAgentsExportJSONWithPrompt(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	runID := "run-export-prompt"
	logDir := ".asagiri/logs/task-f/agents/dev"
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-f", RunID: runID, AgentID: "dev", LogDir: logDir,
	}))
	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "prompt.md"), []byte("full"), 0o644))

	outDir := filepath.Join(dir, "export-prompt")
	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "export", runID, "--output", outDir, "--include-prompt", "--json"})
	require.NoError(t, root.Execute())

	var report agentledger.ExportReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.True(t, report.IncludePrompt)

	var preview agentledger.ReplayPreviewReport
	body, err := os.ReadFile(filepath.Join(outDir, "replay-preview.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &preview))
	for _, a := range preview.Artifacts {
		if a.Name == "prompt.md" {
			require.Equal(t, "full", a.Content)
		}
	}
}

func TestAgentsDiffJSONIdentical(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	logDir := ".asagiri/logs/task-d/agents/dev"
	entry := agentledger.Entry{
		TaskID: "task-d", Feature: "f", AgentID: "dev", LogDir: logDir,
		PromptHash: "same", ContextHash: "same", OutputHash: "same",
	}
	entry.RunID = "run-d1"
	require.NoError(t, agentledger.Append(dir, entry))
	entry.RunID = "run-d2"
	require.NoError(t, agentledger.Append(dir, entry))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "diff", "run-d1", "run-d2", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentledger.DiffReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentledger.DiffReportVersion, report.ReportVersion)
	require.True(t, report.Identical)
}

func TestAgentsDiffJSONNotFound(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "diff", "run-a", "missing", "--json"})
	err := root.Execute()
	require.Error(t, err)
}

func TestAgentsDiffJSONDifferentHashes(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		RunID: "run-h1", TaskID: "t", AgentID: "dev", PromptHash: "a",
	}))
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		RunID: "run-h2", TaskID: "t", AgentID: "dev", PromptHash: "b",
	}))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "diff", "run-h1", "run-h2", "--json"})
	require.NoError(t, root.Execute())

	var report agentledger.DiffReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.False(t, report.Identical)
}

func TestAgentsRunPreviewNotFound(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "run", "no-such-run", "--preview", "--json"})
	err := root.Execute()
	require.Error(t, err)
}

func TestAgentsRunMissingLogsStillOK(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "t", RunID: "run-nologs", AgentID: "dev",
		LogDir: ".asagiri/logs/t/agents/dev",
	}))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "run", "run-nologs", "--json"})
	require.NoError(t, root.Execute())

	var report agentledger.InspectReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	for _, a := range report.Artifacts {
		require.False(t, a.Exists)
	}
}

func TestAgentsRunsJSONEmptyLedger(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "runs", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentledger.Report
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentledger.ReportVersion, report.ReportVersion)
	require.Equal(t, 0, report.Count)
}

func TestAgentsRunsFilterByTask(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	valid := true
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-a", RunID: "run-1", Feature: "feat", AgentID: "dev", Phase: "dev",
		PromptHash: agentledger.HashText("p"), OutputHash: agentledger.HashText("out"),
		ContractValid: &valid, LogDir: ".asagiri/logs/task-a/agents/dev",
	}))
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-b", RunID: "run-2", Feature: "feat", AgentID: "dev", Phase: "dev",
		PromptHash: agentledger.HashText("p2"), OutputHash: agentledger.HashText("out2"), LogDir: ".asagiri/logs/task-b/agents/dev",
	}))

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"agents", "runs", "--task", "task-a", "--json"})
	require.NoError(t, root.Execute())

	var report agentledger.Report
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, 1, report.Count)
	require.Equal(t, "task-a", report.Entries[0].TaskID)
}

func TestAgentsExternalJSON(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "external", "--json"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	var report agentexternal.Report
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentexternal.ReportVersion, report.ReportVersion)
	require.True(t, report.ReadOnly)
	require.NotEmpty(t, report.Policy)
	require.NotEmpty(t, report.Targets)
}

func TestAgentsExternalTextOutput(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"agents", "external"})
	require.NoError(t, root.Execute())
	body := out.String()
	require.Contains(t, body, "Asagiri Agents External")
	require.Contains(t, body, "Politique")
}

func TestAgentsExternalSyncDryRunJSON(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	root := newRootCmd()
	root.SetArgs([]string{"agents", "sync", "--write", "--agent", "dev"})
	require.NoError(t, root.Execute())

	extPath := filepath.Join(dir, "external-dev.md")
	devPath := filepath.Join(dir, ".asagiri", "agents", "dev.yaml")
	data, err := os.ReadFile(devPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(devPath, []byte(string(data)+`
external:
  kind: kiro-agent
  path: `+extPath+`
`), 0o644))

	var out, errOut bytes.Buffer
	root = newRootCmd()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"agents", "external", "sync", "--agent", "dev", "--json"})
	require.NoError(t, root.Execute())
	requireStdoutSingleJSON(t, out.Bytes())
	require.Contains(t, errOut.String(), "asa agents external sync --write")

	var report agentexternal.SyncReport
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, agentexternal.SyncReportVersion, report.ReportVersion)
	require.Equal(t, "check", report.Mode)
	require.False(t, report.Wrote)
	_, err = os.Stat(extPath)
	require.True(t, os.IsNotExist(err))
}

func TestAgentsExternalSyncWriteCreatesFile(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	root := newRootCmd()
	root.SetArgs([]string{"agents", "sync", "--write", "--agent", "dev"})
	require.NoError(t, root.Execute())

	extPath := filepath.Join(dir, "external-dev.md")
	devPath := filepath.Join(dir, ".asagiri", "agents", "dev.yaml")
	data, err := os.ReadFile(devPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(devPath, []byte(string(data)+`
external:
  kind: kiro-agent
  path: `+extPath+`
`), 0o644))

	root = newRootCmd()
	root.SetArgs([]string{"agents", "external", "sync", "--write", "--agent", "dev"})
	require.NoError(t, root.Execute())

	content, err := os.ReadFile(extPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "asagiri: true")
}

func TestAgentsExternalSyncConflictExitNonZero(t *testing.T) {
	dir, _ := trustWorkTestRepo(t)
	root := newRootCmd()
	root.SetArgs([]string{"agents", "sync", "--write", "--agent", "dev"})
	require.NoError(t, root.Execute())

	extPath := filepath.Join(dir, "external-dev.md")
	require.NoError(t, os.WriteFile(extPath, []byte("manual\n"), 0o644))
	devPath := filepath.Join(dir, ".asagiri", "agents", "dev.yaml")
	data, err := os.ReadFile(devPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(devPath, []byte(string(data)+`
external:
  kind: kiro-agent
  path: `+extPath+`
`), 0o644))

	root = newRootCmd()
	root.SetArgs([]string{"agents", "external", "sync", "--write", "--agent", "dev"})
	require.Error(t, root.Execute())
}
