package mission

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func cockpitFixture(width, height int) ViewModel {
	fixed := time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC)
	return ViewModel{
		Workspace:        "asagiri",
		Branch:           "main",
		RuntimeStatus:    "running",
		SessionStatus:    "active",
		ActiveAgents:     []bus.ActiveAgentSummary{{Role: "implementer", AgentRef: "cursor", Status: "running"}},
		Trust:            bus.TrustSummaryResult{Overall: 0.8, Dimensions: []bus.TrustDimensionScore{{Label: "Architecture", Score: 0.82}}},
		Flow:             bus.FlowGraphResult{FlowID: "onboarding", Steps: []bus.FlowGraphStep{{ID: "spec", Label: "spec", Status: "succeeded"}}},
		Runs:             []bus.RunSummary{{ID: "run-1", Feature: "cockpit", Status: "running"}, {ID: "run-2", Feature: "spec-ui", Status: "completed"}},
		Events:           []bus.EventSummary{{Type: "runtime.started", CreatedAt: fixed}},
		EventFeed:        components.EventFeedViewModel{Filter: "all", Search: "(none)"},
		CostTodayEUR:     0.42,
		Now:              fixed,
		DisableAnimations: true,
		Width:            width,
		Height:           height,
		CompactThreshold: 100,
		Theme:            theme.Default(),
	}
}

func TestCockpitResponsiveColumnsFitWidth(t *testing.T) {
	for _, width := range []int{80, 140, 200} {
		vm := cockpitFixture(width, 40)
		got := RenderCockpit(vm)
		require.LessOrEqual(t, lipgloss.Width(got), width, "cockpit must not overflow terminal width %d", width)
		require.NotEmpty(t, got)
		for _, title := range []string{"Runtime", "Trust", "Agents", "Active Flow", "Runs", "Events"} {
			require.Contains(t, stripCockpitANSI(got), title, "pane %q missing at width %d", title, width)
		}
	}
}

func TestCockpitFallsBackToFlatWithoutGeometry(t *testing.T) {
	vm := cockpitFixture(0, 0)
	require.Equal(t, Render(vm), RenderCockpit(vm))
}

func TestRenderRunsSummaryPane(t *testing.T) {
	vm := cockpitFixture(140, 40)
	got := stripCockpitANSI(renderRunsSummaryPane(vm))
	require.Contains(t, got, "cockpit")
	require.Contains(t, got, "running")
	require.Contains(t, got, "✓")

	empty := renderRunsSummaryPane(ViewModel{})
	require.Equal(t, "No recent runs", empty)
}

func TestCockpitGolden(t *testing.T) {
	cases := map[string]ViewModel{
		"cockpit_compact": cockpitFixture(80, 30),
		"cockpit_wide":    cockpitFixture(140, 40),
		"cockpit_ultra":   cockpitFixture(200, 45),
	}
	for name, vm := range cases {
		got := RenderCockpit(vm)
		golden := filepath.Join("testdata", name+".txt")
		if os.Getenv("UPDATE_GOLDEN") == "1" {
			require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
			require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
			continue
		}
		want, err := os.ReadFile(golden)
		if os.IsNotExist(err) {
			t.Fatalf("golden %s missing; run with UPDATE_GOLDEN=1", golden)
		}
		require.NoError(t, err)
		require.Equal(t, string(want), got, "golden mismatch for %s", name)
	}
}

func stripCockpitANSI(v string) string {
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
