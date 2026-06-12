package agentledger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func seedExportRun(t *testing.T, dir, runID string) string {
	t.Helper()
	logDir := ".asagiri/logs/task-exp/agents/dev"
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-exp", RunID: runID, Feature: "feat", AgentID: "dev",
		PromptHash: agentledger.HashText("secret prompt"), LogDir: logDir,
	}))
	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "prompt.md"), []byte("secret prompt"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "context.json"), []byte(`{"task_id":"task-exp"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "invocation.json"), []byte(`{"provider":"exec"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "contract.json"), []byte(`{"valid":true}`), 0o644))
	return logDir
}

func TestExportRunNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := agentledger.Export(dir, "missing", agentledger.ExportOptions{
		OutputDir: filepath.Join(dir, "out"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "introuvable")
}

func TestExportWithoutPromptInPreview(t *testing.T) {
	dir := t.TempDir()
	runID := "run-no-prompt"
	seedExportRun(t, dir, runID)
	outDir := filepath.Join(dir, "bundle")

	report, err := agentledger.Export(dir, runID, agentledger.ExportOptions{
		OutputDir:     outDir,
		IncludePrompt: false,
	})
	require.NoError(t, err)
	require.Equal(t, agentledger.ExportReportVersion, report.ReportVersion)
	require.FileExists(t, filepath.Join(outDir, "manifest.json"))
	require.FileExists(t, filepath.Join(outDir, "artifacts", "prompt.md"))

	var preview agentledger.ReplayPreviewReport
	body, err := os.ReadFile(filepath.Join(outDir, "replay-preview.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &preview))
	for _, a := range preview.Artifacts {
		if a.Name == "prompt.md" {
			require.Empty(t, a.Content)
		}
		if a.Name == "context.json" {
			require.Contains(t, a.Content, "task-exp")
		}
	}
}

func TestExportWithPromptInPreview(t *testing.T) {
	dir := t.TempDir()
	runID := "run-with-prompt"
	seedExportRun(t, dir, runID)
	outDir := filepath.Join(dir, "bundle-prompt")

	_, err := agentledger.Export(dir, runID, agentledger.ExportOptions{
		OutputDir:     outDir,
		IncludePrompt: true,
	})
	require.NoError(t, err)

	var preview agentledger.ReplayPreviewReport
	body, err := os.ReadFile(filepath.Join(outDir, "replay-preview.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &preview))
	for _, a := range preview.Artifacts {
		if a.Name == "prompt.md" {
			require.Equal(t, "secret prompt", a.Content)
		}
	}
}

func TestExportManifestStable(t *testing.T) {
	dir := t.TempDir()
	runID := "run-manifest"
	seedExportRun(t, dir, runID)
	outDir := filepath.Join(dir, "bundle-manifest")

	report, err := agentledger.Export(dir, runID, agentledger.ExportOptions{OutputDir: outDir})
	require.NoError(t, err)

	var manifest agentledger.ExportManifest
	body, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &manifest))
	require.Equal(t, runID, manifest.RunID)
	require.NotEmpty(t, manifest.Files)
	for i := 1; i < len(manifest.Files); i++ {
		require.LessOrEqual(t, manifest.Files[i-1].Path, manifest.Files[i].Path)
	}
	require.NotEmpty(t, manifest.Files[0].SHA256)
	require.Greater(t, manifest.Files[0].SizeBytes, int64(0))

	var foundManifest bool
	for _, f := range report.Files {
		if f.Path == "manifest.json" {
			foundManifest = true
			require.NotEmpty(t, f.SHA256)
		}
	}
	require.True(t, foundManifest)
}
