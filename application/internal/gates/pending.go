package gates

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

// PendingPhase describes what the operator must do before the workflow continues.
type PendingPhase string

const (
	PendingPhaseSubmit PendingPhase = "submit"
	PendingPhaseResume PendingPhase = "resume"
)

// PendingGate is a blocking work gate that is not yet satisfied on a task scope.
// Orthogonal to task business status (implemented, verified, …).
type PendingGate struct {
	Gate     string
	Scope    string
	Blocking bool
	Phase    PendingPhase
}

// BlockingPendingForTask returns the highest-priority blocking gate for a task, if any.
func BlockingPendingForTask(repoRoot string, cfg *config.Config, task sqlite.Task) (PendingGate, bool) {
	if cfg == nil {
		return PendingGate{}, false
	}
	if pg, ok := humanReviewPending(repoRoot, cfg.Work.Gates.HumanReview, task); ok {
		return pg, true
	}
	return PendingGate{}, false
}

// LastGateEntry returns the last history entry for gateName in a task payload JSON.
func LastGateEntry(payloadJSON, gateName string) (asagiri.GateHistoryEntry, bool) {
	task, err := unmarshalTaskPayload(payloadJSON)
	if err != nil {
		return asagiri.GateHistoryEntry{}, false
	}
	return lastGateEntryFromTask(task, gateName)
}

// GateEntrySatisfied reports whether a persisted gate entry clears the pending state.
func GateEntrySatisfied(warnAdvisory bool, entry asagiri.GateHistoryEntry) bool {
	switch strings.ToLower(strings.TrimSpace(entry.Status)) {
	case string(VerdictPass):
		return true
	case string(VerdictWarn):
		return warnAdvisory
	default:
		return false
	}
}

// SubmitCommand returns the CLI command to submit a verdict for this pending gate.
func (p PendingGate) SubmitCommand() string {
	switch p.Gate {
	case HumanReviewGateName:
		return fmt.Sprintf("asa gates submit human_review --task %s --verdict pass", p.Scope)
	default:
		return fmt.Sprintf("asa gates submit %s --task %s --verdict pass", p.Gate, p.Scope)
	}
}

// ResumeCommand returns the CLI command to resume the workflow after gate action.
func (p PendingGate) ResumeCommand() string {
	return "asa continue --yes"
}

// DevResumeCommand returns a direct dev primitive when continue is not desired in tests.
func (p PendingGate) DevResumeCommand(feature, agent string) string {
	if agent == "" {
		agent = "dev"
	}
	return fmt.Sprintf("asa dev %s --task %s --agent %s", feature, p.Scope, agent)
}

// FormatPendingAction renders the canonical gate resume UX block.
func FormatPendingAction(p PendingGate, feature string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Gate %s requires action.\n", p.Gate)
	fmt.Fprintf(&sb, "Next step:\n  %s\n", p.nextStepCommand(feature))
	if p.Phase == PendingPhaseSubmit {
		fmt.Fprintf(&sb, "Then:\n  %s\n", p.ResumeCommand())
	}
	return sb.String()
}

func (p PendingGate) nextStepCommand(feature string) string {
	switch p.Phase {
	case PendingPhaseSubmit:
		return p.SubmitCommand()
	case PendingPhaseResume:
		return p.ResumeCommand()
	default:
		return p.SubmitCommand()
	}
}
