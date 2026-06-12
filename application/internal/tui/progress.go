package tui

// ProgressKind identifies high-level phases for the CLI.
type ProgressKind string

const (
	ProgressInvestigate ProgressKind = "investigate"
	ProgressContext     ProgressKind = "context"
	ProgressEstimate    ProgressKind = "estimate"
	ProgressExecute     ProgressKind = "execute"
)

// PhaseMessage returns a user-facing label.
func PhaseMessage(k ProgressKind) string {
	switch k {
	case ProgressInvestigate:
		return "Investigation locale"
	case ProgressContext:
		return "Optimisation contexte"
	case ProgressEstimate:
		return "Estimation coût"
	case ProgressExecute:
		return "Exécution"
	default:
		return string(k)
	}
}
