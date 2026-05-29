package workflow

import (
	"errors"
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const DefaultResumeMaxSteps = 20

var (
	ErrInvalidTransition       = errors.New("transition d'état invalide")
	ErrStepAlreadyDone         = errors.New("étape déjà réussie — utilisez --force pour relancer")
	ErrResumeMaxStepsExceeded  = errors.New("resume: max steps exceeded with remaining work")
	ErrResumeGateBlocked       = errors.New("resume: step blocked by gate")
)

// allowedTransitions maps from-status to permitted next statuses (spec §12).
var allowedTransitions = map[string][]string{
	asagiri.StatusPending:      {asagiri.StatusPlanned, asagiri.StatusAborted},
	asagiri.StatusPlanned:      {asagiri.StatusEnriched, asagiri.StatusAborted},
	asagiri.StatusEnriched:     {asagiri.StatusRunning, asagiri.StatusAborted},
	asagiri.StatusRunning:      {asagiri.StatusImplemented, asagiri.StatusFailed, asagiri.StatusAborted},
	asagiri.StatusImplemented:  {asagiri.StatusVerified, asagiri.StatusVerifyFailed, asagiri.StatusAborted},
	asagiri.StatusVerifyFailed: {asagiri.StatusImplemented, asagiri.StatusAborted},
	asagiri.StatusVerified:     {asagiri.StatusReviewed, asagiri.StatusReviewFailed, asagiri.StatusAborted},
	asagiri.StatusReviewFailed: {asagiri.StatusVerified, asagiri.StatusAborted},
	asagiri.StatusReviewed:     {asagiri.StatusReadyForPR, asagiri.StatusAborted},
	asagiri.StatusReadyForPR:   {asagiri.StatusMerged, asagiri.StatusAborted},
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
		return asagiri.StatusImplemented
	case "failed":
		return asagiri.StatusFailed
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
	case asagiri.StatusPending:
		return 0
	case asagiri.StatusPlanned:
		return 1
	case asagiri.StatusEnriched:
		return 2
	case asagiri.StatusRunning, asagiri.StatusFailed:
		return 3
	case asagiri.StatusImplemented, asagiri.StatusVerifyFailed:
		return 4
	case asagiri.StatusVerified, asagiri.StatusReviewFailed:
		return 5
	case asagiri.StatusReviewed:
		return 6
	case asagiri.StatusReadyForPR, asagiri.StatusMerged:
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
