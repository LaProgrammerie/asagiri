package doctor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSummarizeModeGate(t *testing.T) {
	active := summarizeModeGate("enrich", true, config.GovernanceModePerTask, true, false)
	require.Equal(t, "active", active.Status)

	inactive := summarizeModeGate("enrich", true, "smart", false, true)
	require.Equal(t, "invalid_mode", inactive.Status)

	off := summarizeModeGate("enrich", true, config.GovernanceModeOff, false, false)
	require.Equal(t, "inactive", off.Status)
}

func TestFormatTextSnapshot(t *testing.T) {
	report := sampleReport()
	Finalize(&report)
	var buf bytes.Buffer
	require.NoError(t, FormatText(&buf, report, false))
	assertGolden(t, "doctor_ready.txt", buf.String())
}

func TestFormatTextWarnOnly(t *testing.T) {
	report := sampleReport()
	report.Checks = append(report.Checks, Check{ID: "docs.product", Status: StatusWarn, Message: "placeholder"})
	Finalize(&report)
	var buf bytes.Buffer
	require.NoError(t, FormatText(&buf, report, false))
	require.Contains(t, buf.String(), "avertissements")
}

func TestFormatTextStrictWarnFails(t *testing.T) {
	report := sampleReport()
	report.Checks = append(report.Checks, Check{ID: "docs.product", Status: StatusWarn, Message: "placeholder"})
	Finalize(&report)
	var buf bytes.Buffer
	require.Error(t, FormatText(&buf, report, true))
}

func TestFormatJSONRoundTrip(t *testing.T) {
	report := sampleReport()
	Finalize(&report)
	var buf bytes.Buffer
	require.NoError(t, FormatJSON(&buf, report))
	require.Contains(t, buf.String(), `"report_version": "doctor-v1"`)
	require.Contains(t, buf.String(), `"ready": true`)
	require.Contains(t, buf.String(), `"warnings": []`)
	require.Contains(t, buf.String(), `"failures": []`)
}

func sampleReport() Report {
	return Report{
		ReportVersion: ReportVersion,
		Ready:         true,
		Repository: RepositoryInfo{
			GitRoot:      "/repo",
			ConfigPath:   "/repo/.asagiri/config.yaml",
			ConfigLoaded: true,
		},
		State: StateInfo{
			SQLitePath:    "/repo/.asagiri/state.sqlite",
			SQLitePresent: true,
			SchemaVersion: 1,
			RunCount:      2,
			TaskCount:     5,
			ActiveFeature: "onboarding",
		},
		Gates: []GateInfo{
			{Name: "plan", Status: "disabled", Detail: "gate désactivée"},
			{Name: "enrich", Enabled: true, Mode: "per-task", Status: "active", Detail: "mode per-task"},
		},
		Agents: []AgentInfo{
			{Role: "dev", LogicalID: "developer", Command: "cursor", Status: "ok", InPath: true},
		},
		Trust: &TrustInfo{
			Feature: "onboarding", Verdict: "Acceptable", TasksAtRisk: 1, Summary: "1 task à risque",
		},
		Checks: []Check{
			{ID: "git", Status: "ok"},
			{ID: "config", Status: "ok"},
		},
		NextActions: []Action{{Title: "Continuer le workflow", CLI: "asa next --feature onboarding"}},
	}
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o644))
		return
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden %s missing; run with UPDATE_GOLDEN=1", goldenPath)
	}
	require.Equal(t, string(want), got)
}
