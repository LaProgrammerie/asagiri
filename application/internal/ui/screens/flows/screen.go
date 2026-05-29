package flows

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// ViewModel contains flow explorer data.
type ViewModel struct {
	Flow    bus.FlowExplorerResult
	ShowCLI bool
}

// Render returns flow explorer content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Flow: " + value(vm.Flow.FlowID, "-") + "\n")
	if len(vm.Flow.Steps) == 0 {
		b.WriteString("- none\n")
		return strings.TrimRight(b.String(), "\n")
	}
	for i, step := range vm.Flow.Steps {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("%s %s (%s)\n", statusGlyph(step.Status), value(step.Label, step.ID), value(step.Status, "unknown")))
		b.WriteString(fmt.Sprintf("  API: %s\n", value(step.API, "n/a")))
		b.WriteString(fmt.Sprintf("  Service: %s\n", value(step.Service, "n/a")))
		b.WriteString(fmt.Sprintf("  Event: %s\n", value(step.Event, "n/a")))
		b.WriteString(fmt.Sprintf("  Trust: %.0f%%  Risk: %s\n", step.TrustScore*100, value(step.Risk, "unknown")))
		if len(step.Tests) > 0 {
			b.WriteString("  Tests: " + strings.Join(step.Tests, ", ") + "\n")
		}
		if len(step.Metrics) > 0 {
			b.WriteString("  Metrics: " + strings.Join(step.Metrics, ", ") + "\n")
		}
		if vm.ShowCLI {
			b.WriteString(fmt.Sprintf("  CLI: asa flow open %s\n", value(vm.Flow.FlowID, "<flow>")))
		}
		if i >= 4 {
			break
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func statusGlyph(status string) string {
	switch status {
	case "succeeded", "completed", "done":
		return "✓"
	case "running":
		return "⠋"
	case "failed":
		return "✕"
	default:
		return "○"
	}
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
