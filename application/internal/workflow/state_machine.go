package workflow

import (
	"errors"
	"fmt"

	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
)

var (
	ErrInvalidTransition = errors.New("transition d'état invalide")
	ErrStepAlreadyDone   = errors.New("étape déjà réussie — utilisez --force pour relancer")
)

// allowedTransitions maps from-status to permitted next statuses (spec §12).
var allowedTransitions = map[string][]string{
	agentflow.StatusPending:      {agentflow.StatusPlanned, agentflow.StatusAborted},
	agentflow.StatusPlanned:      {agentflow.StatusEnriched, agentflow.StatusAborted},
	agentflow.StatusEnriched:     {agentflow.StatusRunning, agentflow.StatusAborted},
	agentflow.StatusRunning:      {agentflow.StatusImplemented, agentflow.StatusFailed, agentflow.StatusAborted},
	agentflow.StatusImplemented:  {agentflow.StatusVerified, agentflow.StatusVerifyFailed, agentflow.StatusAborted},
	agentflow.StatusVerifyFailed: {agentflow.StatusImplemented, agentflow.StatusAborted},
	agentflow.StatusVerified:     {agentflow.StatusReviewed, agentflow.StatusReviewFailed, agentflow.StatusAborted},
	agentflow.StatusReviewFailed: {agentflow.StatusVerified, agentflow.StatusAborted},
	agentflow.StatusReviewed:     {agentflow.StatusReadyForPR, agentflow.StatusAborted},
	agentflow.StatusReadyForPR:   {agentflow.StatusMerged, agentflow.StatusAborted},
}

// TransitionTask moves a task to toStatus unless already at a terminal success state.
func TransitionTask(from, to string, force bool) error {
	from = normalizeStatus(from)
	to = normalizeStatus(to)
	if from == to {
		if force {
			return nil
		}
		return fmt.Errorf("%w: %s", ErrStepAlreadyDone, from)
	}
	if !canTransition(from, to) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}

func normalizeStatus(s string) string {
	switch s {
	case "done":
		return agentflow.StatusImplemented
	case "failed":
		return agentflow.StatusFailed
	default:
		return s
	}
}

func canTransition(from, to string) bool {
	next, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, n := range next {
		if n == to {
			return true
		}
	}
	return false
}

// stepRank orders pipeline progress (lower = earlier).
func stepRank(status string) int {
	switch normalizeStatus(status) {
	case agentflow.StatusPending:
		return 0
	case agentflow.StatusPlanned:
		return 1
	case agentflow.StatusEnriched:
		return 2
	case agentflow.StatusRunning, agentflow.StatusFailed:
		return 3
	case agentflow.StatusImplemented, agentflow.StatusVerifyFailed:
		return 4
	case agentflow.StatusVerified, agentflow.StatusReviewFailed:
		return 5
	case agentflow.StatusReviewed:
		return 6
	case agentflow.StatusReadyForPR, agentflow.StatusMerged:
		return 7
	default:
		return 0
	}
}

// NextWorkflowStep returns the CLI step name to resume for a run based on task statuses.
func NextWorkflowStep(taskStatuses []string) string {
	if len(taskStatuses) == 0 {
		return "plan"
	}
	minRank := 7
	for _, st := range taskStatuses {
		r := stepRank(st)
		if r < minRank {
			minRank = r
		}
	}
	switch minRank {
	case 0:
		return "plan"
	case 1:
		return "enrich"
	case 2, 3:
		return "dev"
	case 4:
		return "verify"
	case 5:
		return "review"
	case 6:
		return "report"
	default:
		return ""
	}
}
