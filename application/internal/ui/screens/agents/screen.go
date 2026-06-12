package agents

import (
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// ViewModel contains agent theatre data.
type ViewModel struct {
	Theatre bus.AgentTheatreResult
	ShowCLI bool
}

// Render returns the live agents screen content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Agents Watch\n")
	if vm.Theatre.Warning != "" {
		b.WriteString("Warning: " + vm.Theatre.Warning + "\n")
	}
	if len(vm.Theatre.Agents) == 0 {
		b.WriteString("- no active agents")
		return b.String()
	}
	for i, agent := range vm.Theatre.Agents {
		if i >= 8 {
			break
		}
		fmt.Fprintf(&b, "\n[%s] %s (%s)\n", value(agent.Role, "agent"), value(agent.AgentRef, "-"), value(agent.Status, "unknown"))
		fmt.Fprintf(&b, "Task: %s\n", value(agent.Task, "-"))
		if agent.FilesActive > 0 {
			fmt.Fprintf(&b, "Files: %d\n", agent.FilesActive)
		}
		if agent.Hypothesis != "" {
			b.WriteString("Hypothesis: " + agent.Hypothesis + "\n")
		}
		if agent.TokensEstimated > 0 || agent.CostEUR > 0 {
			fmt.Fprintf(&b, "Tokens: %d  Cost: €%.2f\n", agent.TokensEstimated, agent.CostEUR)
		}
		if agent.Duration > 0 {
			b.WriteString("Duration: " + formatDuration(agent.Duration) + "\n")
		}
		if agent.LastOutput != "" {
			b.WriteString("Last output: " + agent.LastOutput + "\n")
		}
		if agent.Confidence > 0 {
			fmt.Fprintf(&b, "Confidence: %.0f%%\n", agent.Confidence*100)
		}
		if vm.ShowCLI {
			b.WriteString("CLI: asa agents watch\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return d.Round(10 * time.Millisecond).String()
	}
	return d.Round(time.Second).String()
}
