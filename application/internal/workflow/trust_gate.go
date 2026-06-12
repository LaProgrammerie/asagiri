package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const trustGateName = "trust"

func (s *Service) processTrustGate(ctx context.Context, feature string, task sqlite.Task) error {
	if s.cfg == nil || !s.cfg.Work.Gates.Trust.IsActive() {
		return nil
	}
	_ = ctx

	var result gates.Result
	if s.dryRun {
		result = s.gateDryRunResult("trust_gate", "trust_gate", task.ID, "trust gate dry-run: simulated pass", nil)
	} else {
		report, err := worktrust.BuildTaskReport(s.repoRoot, s.cfg, task)
		if err != nil {
			result = gates.Result{
				GateID:     "trust_gate",
				GateType:   "trust_gate",
				Scope:      task.ID,
				Status:     gates.VerdictFail,
				Confidence: 0,
				Notes:      []string{fmt.Sprintf("trust synthesis failed: %v", err)},
			}
		} else {
			result = worktrust.WorkTrustReportToGateResult(report, s.cfg.Work.Gates.Trust)
		}
	}
	return s.persistTrustGateVerdict(feature, task, result)
}

func (s *Service) persistTrustGateVerdict(feature string, task sqlite.Task, v gates.Result) error {
	at := time.Now().UTC().Format(time.RFC3339)
	entry := gateHistoryEntryFromResult(trustGateName, v, at, 0)

	canonical, err := payloadToCanonical(task.PayloadJSON)
	if err != nil {
		return err
	}
	if canonical.Gates == nil {
		canonical.Gates = &asagiri.TaskGates{}
	}
	canonical.Gates.History = append(canonical.Gates.History, entry)
	canonical.TouchMetadata(time.Now().UTC())

	payload, err := canonicalToPayload(canonical)
	if err != nil {
		return err
	}
	if err := s.store.UpdateTask(&sqlite.Task{ID: task.ID, PayloadJSON: payload}); err != nil {
		return err
	}

	return s.persistGateLogs(
		task.ID, "task", trustGateName, feature, "worktrust",
		"trust_gate", "Trust gate", "", v,
	)
}
