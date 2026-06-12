package coordination

import (
	"fmt"
	"strings"
)

const defaultRuntimeTitle = "Asagiri Multi-Agent Runtime"

// RuntimeDisplayStatus is the lifecycle state of a pipeline role in terminal UX (spec-my-D §19).
type RuntimeDisplayStatus string

const (
	RuntimeDisplayCompleted RuntimeDisplayStatus = "completed"
	RuntimeDisplayRunning   RuntimeDisplayStatus = "running"
	RuntimeDisplayPending   RuntimeDisplayStatus = "pending"
	RuntimeDisplayFailed    RuntimeDisplayStatus = "failed"
)

// RuntimeDisplayStep is one role line in the §19 terminal snapshot.
type RuntimeDisplayStep struct {
	Role   AgentRole
	Status RuntimeDisplayStatus
}

// ActiveAgent maps a pipeline role to the configured agent ref for terminal display.
type ActiveAgent struct {
	Role     AgentRole
	AgentRef string
}

// MultiAgentRuntimeView feeds FormatMultiAgentRuntime (spec-my-D §19).
type MultiAgentRuntimeView struct {
	Title      string
	Pipeline   []RuntimeDisplayStep
	Agents     []ActiveAgent
	CurrentEUR float64
	BudgetEUR  float64
	Warnings   []string
}

// FormatMultiAgentRuntime renders the §19 terminal snapshot (pipeline, agents, costs, warnings).
func FormatMultiAgentRuntime(v MultiAgentRuntimeView) string {
	title := strings.TrimSpace(v.Title)
	if title == "" {
		title = defaultRuntimeTitle
	}

	var b strings.Builder
	b.WriteString(title)
	b.WriteByte('\n')
	b.WriteString(strings.Repeat("═", len(title)))
	b.WriteByte('\n')

	b.WriteString("Pipeline\n")
	b.WriteString("────────\n")
	for _, step := range v.Pipeline {
		b.WriteString(runtimeDisplayIcon(step.Status))
		b.WriteByte(' ')
		b.WriteString(string(step.Role))
		b.WriteByte('\n')
	}

	b.WriteString("Agents\n")
	b.WriteString("──────\n")
	for _, ag := range v.Agents {
		_, _ = fmt.Fprintf(&b, "%s: %s\n", ag.Role, ag.AgentRef)
	}

	b.WriteString("Costs\n")
	b.WriteString("─────\n")
	_, _ = fmt.Fprintf(&b, "Current: %.2f€\n", v.CurrentEUR)
	_, _ = fmt.Fprintf(&b, "Budget:  %.2f€\n", v.BudgetEUR)

	if len(v.Warnings) > 0 {
		b.WriteString("Warnings\n")
		b.WriteString("────────\n")
		for _, w := range v.Warnings {
			b.WriteString("- ")
			b.WriteString(w)
			b.WriteByte('\n')
		}
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func runtimeDisplayIcon(status RuntimeDisplayStatus) string {
	switch status {
	case RuntimeDisplayCompleted:
		return "✓"
	case RuntimeDisplayRunning:
		return "⠋"
	case RuntimeDisplayFailed:
		return "✗"
	default:
		return "○"
	}
}
