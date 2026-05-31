package runs

import (
	"os"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

func stripANSI(v string) string {
	var out []rune
	in := false
	for _, r := range v {
		if r == '\x1b' {
			in = true
			continue
		}
		if in {
			if r == 'm' {
				in = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

func sampleVM(width int) ViewModel {
	return ViewModel{
		Runs: []bus.RunSummary{
			{ID: "run-1", Feature: "cockpit", Status: "running"},
			{ID: "run-2", Feature: "spec-ui", Status: "completed"},
		},
		Readiness: bus.ReadinessResult{Ready: true, Score: 100},
		Detail: bus.RunDetail{
			ID:       "run-1",
			Feature:  "cockpit",
			Status:   "running",
			Worktree: ".asagiri/worktrees/run-1",
			Pipeline: []bus.RunPipelineStep{
				{ID: "spec", Label: "spec", Status: "done"},
				{ID: "verify", Label: "verify", Status: "running"},
			},
			Validation: "verify: running",
			TrustGate:  bus.TrustSummaryResult{Overall: 0.8},
			CostEUR:    1.42,
			Agents:     []bus.ActiveAgentSummary{{Role: "implementer", AgentRef: "cursor", Status: "running"}},
			Events:     []bus.EventSummary{{Type: "runtime.started"}},
		},
		Model:  NewModel(),
		Width:  width,
		Height: 30,
		Theme:  theme.Default(),
	}
}

func TestRenderRunsListAndDetail(t *testing.T) {
	got := stripANSI(Render(sampleVM(120)))
	require.Contains(t, got, "cockpit")
	require.Contains(t, got, "spec-ui")
	require.Contains(t, got, "Worktree: .asagiri/worktrees/run-1")
	require.Contains(t, got, "spec")
	require.Contains(t, got, "verify")
	require.Contains(t, got, "Validation: verify: running")
	require.Contains(t, got, "€1.42")
	require.Contains(t, got, "cursor")
}

func TestRenderRunsEmptyState(t *testing.T) {
	got := stripANSI(Render(ViewModel{Theme: theme.Default(), Width: 120, Height: 30}))
	require.Contains(t, got, "No runs yet")
	require.Contains(t, got, "asa onboard --ui")
}

func TestRenderRunsEmptyWhenNotOnboarded(t *testing.T) {
	got := stripANSI(Render(ViewModel{
		Runs:      []bus.RunSummary{{ID: "run-1", Feature: "x", Status: "running"}},
		Readiness: bus.ReadinessResult{Ready: false, Score: 20},
		Theme:     theme.Default(),
		Width:     120,
		Height:    30,
	}))
	require.Contains(t, got, "No runs yet")
	require.Contains(t, got, "asa onboard --ui")
}

func TestRunsSelectionClamps(t *testing.T) {
	m := NewModel()
	m.SelectIndex(5, 2)
	require.Equal(t, 1, m.Cursor)
	require.Equal(t, "run-2", m.SelectedRunID([]bus.RunSummary{{ID: "run-1"}, {ID: "run-2"}}))

	m.SelectIndex(-3, 2)
	require.Equal(t, 0, m.Cursor)
	require.Equal(t, "", m.SelectedRunID(nil))
}

func TestRenderRunsNarrowStacks(t *testing.T) {
	got := stripANSI(Render(sampleVM(60)))
	require.Contains(t, got, "Runs")
	require.Contains(t, got, "Run detail")
}

func TestRenderRunsGolden(t *testing.T) {
	got := Render(sampleVM(120))
	golden := "testdata/runs_wide.txt"
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
		return
	}
	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) {
		t.Fatalf("golden %s missing; run with UPDATE_GOLDEN=1", golden)
	}
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}
