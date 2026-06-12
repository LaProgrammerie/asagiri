package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const humanReviewAgentLabel = "human"

var humanReviewParseConfig = gates.ParseConfig{
	BlockKey:          "human_review",
	MissingBlockError: "human_review block missing from verdict file",
	ParseErrorNote:    "human_review_parse_error",
}

func (s *Service) humanReviewVerdictPath(taskID string) string {
	name := ""
	if s.cfg != nil {
		name = s.cfg.Work.Gates.HumanReview.VerdictFile
	}
	return gates.HumanReviewVerdictPath(s.repoRoot, taskID, name)
}

func humanReviewEvidenceRefs(feature string, task sqlite.Task) []gates.EvidenceRef {
	refs := []gates.EvidenceRef{
		{Kind: "task", Path: task.ID, Note: "task awaiting human review"},
		{Kind: "feature", Path: feature, Note: "feature scope"},
	}
	if wt := strings.TrimSpace(task.WorktreePath); wt != "" {
		refs = append(refs, gates.EvidenceRef{
			Kind: "worktree",
			Path: wt,
			Note: "implementation worktree",
		})
	}
	return refs
}

func readHumanReviewVerdictFile(path, taskID string) (gates.Result, string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			pg := gates.PendingGate{Gate: gates.HumanReviewGateName, Scope: taskID, Blocking: true, Phase: gates.PendingPhaseSubmit}
			return gates.Result{}, "", fmt.Errorf(
				"human review gate: verdict file missing at %s\n%s",
				path, gates.FormatPendingAction(pg, ""),
			)
		}
		return gates.Result{}, "", fmt.Errorf("human review gate: read verdict file: %w", err)
	}
	content := string(body)
	parsed := gates.ParseResult(content, humanReviewParseConfig)
	parsed.GateID = gates.HumanReviewGateName
	parsed.GateType = gates.HumanReviewGateName
	return parsed, content, nil
}

func (s *Service) runHumanReviewGate(feature string, task sqlite.Task) (gates.Result, string, error) {
	evidence := humanReviewEvidenceRefs(feature, task)
	if s.dryRun {
		return s.gateDryRunResult(gates.HumanReviewGateName, gates.HumanReviewGateName, task.ID, "human review dry-run: simulated pass", evidence), "", nil
	}

	path := s.humanReviewVerdictPath(task.ID)
	parsed, raw, err := readHumanReviewVerdictFile(path, task.ID)
	if err != nil {
		return gates.Result{}, "", err
	}
	parsed.Scope = task.ID
	parsed.Evidence = evidence
	return gates.ClassifyResult(parsed, s.cfg.Work.Gates.HumanReview.FailOn), raw, nil
}

func (s *Service) persistHumanReviewVerdict(feature string, task sqlite.Task, v gates.Result, verdictRaw string) error {
	at := time.Now().UTC().Format(time.RFC3339)
	entry := gateHistoryEntryFromResult(gates.HumanReviewGateName, v, at, 0)

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

	stdout := verdictRaw
	if strings.TrimSpace(stdout) == "" {
		stdout = fmt.Sprintf("(human review verdict at %s)", s.humanReviewVerdictPath(task.ID))
	}
	return s.persistGateLogs(
		task.ID, "task", gates.HumanReviewGateName, feature, humanReviewAgentLabel,
		"human_review", "Human review gate", stdout, v,
	)
}

func lastHumanReviewEntry(task sqlite.Task) (asagiri.GateHistoryEntry, bool) {
	return gates.LastGateEntry(task.PayloadJSON, gates.HumanReviewGateName)
}

func humanReviewEntrySatisfied(warnAdvisory bool, entry asagiri.GateHistoryEntry) bool {
	return gates.GateEntrySatisfied(warnAdvisory, entry)
}

func (s *Service) processHumanReviewAfterDev(ctx context.Context, feature string, task sqlite.Task) error {
	if s.cfg == nil || !s.cfg.Work.Gates.HumanReview.IsActive() {
		return nil
	}
	_ = ctx

	if fresh, err := s.store.GetTask(task.ID); err == nil {
		task = *fresh
	}
	if entry, ok := lastHumanReviewEntry(task); ok && humanReviewEntrySatisfied(s.cfg.Work.Gates.HumanReview.WarnAdvisory(), entry) {
		return nil
	}

	verdict, raw, runErr := s.runHumanReviewGate(feature, task)
	if runErr != nil {
		return runErr
	}
	if err := s.persistHumanReviewVerdict(feature, task, verdict, raw); err != nil {
		return err
	}

	switch verdict.Status {
	case gates.VerdictPass:
		return nil
	case gates.VerdictWarn:
		if s.cfg.Work.Gates.HumanReview.WarnAdvisory() {
			return nil
		}
		if err := s.transitionTask(task, asagiri.StatusFailed, true); err != nil {
			return err
		}
		return fmt.Errorf("human review gate warn (non-advisory): %s", gates.FormatFailure(verdict))
	case gates.VerdictFail:
		if err := s.transitionTask(task, asagiri.StatusFailed, true); err != nil {
			return err
		}
		return fmt.Errorf("human review gate failed: %s", gates.FormatFailure(verdict))
	default:
		return fmt.Errorf("human review gate unknown status %q", verdict.Status)
	}
}

// WriteHumanReviewVerdictFile writes a human review verdict YAML for a task (CLI helper).
func WriteHumanReviewVerdictFile(repoRoot, taskID, verdictFile, verdict string, notes []string, fileSrc string) (string, error) {
	path := gates.HumanReviewVerdictPath(repoRoot, taskID, verdictFile)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if strings.TrimSpace(fileSrc) != "" {
		body, err := os.ReadFile(fileSrc)
		if err != nil {
			return "", fmt.Errorf("read verdict file: %w", err)
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			return "", err
		}
		return path, nil
	}
	v := strings.TrimSpace(strings.ToLower(verdict))
	if v != "pass" && v != "warn" && v != "fail" {
		return "", fmt.Errorf("verdict must be pass, warn, or fail")
	}
	var sb strings.Builder
	sb.WriteString("human_review:\n")
	sb.WriteString("  status: ")
	sb.WriteString(v)
	sb.WriteString("\n  confidence: 1.0\n")
	if len(notes) > 0 {
		sb.WriteString("  notes:\n")
		for _, n := range notes {
			sb.WriteString("    - ")
			sb.WriteString(n)
			sb.WriteString("\n")
		}
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
