package workflow

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

// blockingGateWorkflowError returns an actionable error when a task has a blocking pending gate.
func (s *Service) blockingGateWorkflowError(feature string, task sqlite.Task) error {
	if pg, ok := gates.BlockingPendingForTask(s.repoRoot, s.cfg, task); ok {
		return fmt.Errorf("%s", gates.FormatPendingAction(pg, feature))
	}
	return nil
}

func (s *Service) workflowStepBlockedError(feature string, task sqlite.Task, step string) error {
	if err := s.blockingGateWorkflowError(feature, task); err != nil {
		return err
	}
	if step == "review" && gates.VerifyEvidenceGateBlocksReview(s.cfg, task.Status, task.PayloadJSON) {
		return fmt.Errorf("verify evidence gate required before review: run asa verify %s --task %s --force", feature, task.ID)
	}
	if step == "review" && gates.TrustGateBlocksReview(s.cfg, task.Status, task.PayloadJSON) {
		return fmt.Errorf("trust gate required before review: run asa verify %s --task %s --force", feature, task.ID)
	}
	return nil
}
