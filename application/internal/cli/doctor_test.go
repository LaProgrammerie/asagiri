package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/doctorcli"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestDoctorJSONAfterInit(t *testing.T) {
	_, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateRun(&sqlite.Run{
		ID: "run-doc", Feature: "myfeat", Status: sqlite.StatusRunning, StepsJSON: `[]`,
	}))
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-doc", RunID: "run-doc", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doctor", "--json"})
	require.NoError(t, root.Execute())

	report, err := doctorcli.DecodeJSON(out.Bytes())
	require.NoError(t, err)
	require.Equal(t, "doctor-v1", report.ReportVersion)
	require.True(t, report.Ready)
	require.Empty(t, report.Failures)
	require.True(t, report.Repository.ConfigLoaded)
	require.True(t, report.State.SQLitePresent)
	require.GreaterOrEqual(t, report.State.RunCount, 1)
	require.NotEmpty(t, report.Gates)
	require.NotEmpty(t, report.Agents)
}

func TestDoctorTextSections(t *testing.T) {
	trustWorkTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doctor"})
	require.NoError(t, root.Execute())
	body := out.String()
	require.Contains(t, body, "Asagiri Doctor")
	require.Contains(t, body, "Repository")
	require.Contains(t, body, "Gates (work)")
	require.Contains(t, body, "Agents")
	require.Contains(t, body, "Prochaines actions")
}

func TestDoctorFailExitsNonZero(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"doctor"})
	require.Error(t, root.Execute())
}

func TestDoctorWarnOnlyExitsZero(t *testing.T) {
	doctorFullTestRepo(t)

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doctor", "--full", "--json"})
	require.NoError(t, root.Execute())

	report, err := doctorcli.DecodeJSON(out.Bytes())
	require.NoError(t, err)
	require.True(t, report.Ready)
	require.NotEmpty(t, report.Warnings)
	require.Empty(t, report.Failures)
}

func TestDoctorStrictWarnExitsNonZero(t *testing.T) {
	doctorFullTestRepo(t)

	root := newRootCmd()
	root.SetArgs([]string{"doctor", "--full", "--strict"})
	require.Error(t, root.Execute())
}

func TestDoctorStrictWarnJSONExitsNonZero(t *testing.T) {
	doctorFullTestRepo(t)

	var out, errOut bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SilenceUsage = true
	root.SilenceErrors = true
	root.SetArgs([]string{"doctor", "--full", "--strict", "--json"})
	err := root.Execute()
	require.Error(t, err)

	requireStdoutSingleJSON(t, out.Bytes())
	report, decErr := doctorcli.DecodeJSON(out.Bytes())
	require.NoError(t, decErr)
	require.True(t, report.Ready)
	require.NotEmpty(t, report.Warnings)
}

func TestDoctorStrictWarnJSONSaveExitsNonZero(t *testing.T) {
	doctorFullTestRepo(t)

	root, out, errOut := rootWithSplitIO()
	root.SilenceUsage = true
	root.SilenceErrors = true
	root.SetArgs([]string{"doctor", "--full", "--strict", "--json", "--save"})
	err := root.Execute()
	require.Error(t, err)

	requireStdoutSingleJSON(t, out.Bytes())
	report, decErr := doctorcli.DecodeJSON(out.Bytes())
	require.NoError(t, decErr)
	require.True(t, report.Ready)
	require.NotEmpty(t, report.Warnings)
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/doctor/latest.json")
}

func TestDoctorSave(t *testing.T) {
	repo, store := trustWorkTestRepo(t)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-dsave", RunID: "run-1", Feature: "myfeat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`,
	}))
	_ = store.Close()

	root, out, errOut := rootWithSplitIO()
	root.SetArgs([]string{"doctor", "--save"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "Asagiri Doctor")
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/doctor/latest.json")

	abs := filepath.Join(repo, ".asagiri", "reports", "doctor", "latest.json")
	raw, err := os.ReadFile(abs)
	require.NoError(t, err)
	requireStdoutSingleJSON(t, raw)
	report, err := doctorcli.DecodeJSON(raw)
	require.NoError(t, err)
	require.Equal(t, "doctor-v1", report.ReportVersion)
}

func TestDoctorJSONSave(t *testing.T) {
	trustWorkTestRepo(t)

	root, out, errOut := rootWithSplitIO()
	root.SetArgs([]string{"doctor", "--json", "--save"})
	require.NoError(t, root.Execute())

	requireStdoutSingleJSON(t, out.Bytes())
	report, err := doctorcli.DecodeJSON(out.Bytes())
	require.NoError(t, err)
	require.Equal(t, "doctor-v1", report.ReportVersion)
	requireReportSavedOnStderr(t, errOut.String(), ".asagiri/reports/doctor/latest.json")
}

func TestDoctorArchitectureJSON(t *testing.T) {
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
	root.SetArgs([]string{"doctor", "architecture", "--json"})
	require.NoError(t, root.Execute())

	report, err := doctorcli.DecodeArchitectureJSON(out.Bytes())
	require.NoError(t, err)
	require.Equal(t, "doctor-architecture-v1", report.ReportVersion)
	wantRoot, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	gotRoot, err := filepath.EvalSymlinks(report.Repository.GitRoot)
	require.NoError(t, err)
	require.Equal(t, wantRoot, gotRoot)
}

func TestDoctorSaveWithoutRuntime(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"doctor", "--save"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), ".asagiri")
}

// doctorFullTestRepo : init + gitignore/kiro OK, docs placeholder → WARN onboarding sans FAIL.
func doctorFullTestRepo(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"),
		[]byte(".asagiri/state.sqlite\n.asagiri/worktrees/\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".kiro", "specs", "demo-feature"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs", "ai"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "ai", "01-product.md"),
		[]byte("# Product\n\nTemplate placeholder — à compléter.\n"), 0o644))

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())
}
