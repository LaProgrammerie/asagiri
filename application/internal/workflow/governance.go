package workflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentresolve"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

const governancePromptTemplate = `Tu es un validateur de gouvernance post-implémentation. Tu ne produis AUCUN code et tu ne modifies aucun fichier.
Analyse si la tâche implémentée respecte la spec et le diff fourni.
Réponds UNIQUEMENT avec un bloc YAML de la forme:

governance:
  status: pass|warn|fail
  confidence: 0.0-1.0
  notes:
    - ...
  findings:
    - code: spec_drift|architecture_violation|unexpected_design_change|other
      severity: warn|fail
      message: ...
      actions:
        - ...

--- SPEC ---
%s

--- TASK ---
%s

--- DIFF ---
%s
%s
`

const governanceGateName = "governance"

func (s *Service) runGovernanceGate(ctx context.Context, feature string, task sqlite.Task, worktreePath string) (gates.Result, string, error) {
	evidence := governanceEvidenceRefs(feature, task, worktreePath, s.repoRoot)
	if s.dryRun {
		return s.gateDryRunResult("governance", "governance", task.ID, "governance dry-run: simulated pass", evidence), "", nil
	}

	legacyPrompt, err := s.buildGovernancePrompt(feature, task, worktreePath)
	if err != nil {
		return gates.Result{}, "", err
	}
	agentKey := s.cfg.GovernanceAgent()
	canonical, _ := payloadToCanonical(task.PayloadJSON)
	contextFiles := s.contextFilesForTask(feature, canonical)
	prompt, err := s.resolveGatePrompt(agentresolve.PhaseGovernance, agentKey, feature, task.ID, "", legacyPrompt, contextFiles)
	if err != nil {
		return gates.Result{}, "", err
	}

	stdout, err := s.executeGateAgent(ctx, agentKey, feature, task.ID, worktreePath, prompt, s.governanceAgentHook)
	if err != nil {
		return gates.Result{}, stdout, err
	}

	parsed := parseGovernanceVerdict(stdout)
	parsed.GateID = "governance"
	parsed.GateType = "governance"
	parsed.Scope = task.ID
	parsed.Evidence = evidence
	return gates.ClassifyResult(parsed, s.cfg.Work.Gates.Governance.FailOn), stdout, nil
}

func governanceEvidenceRefs(feature string, task sqlite.Task, worktreePath, repoRoot string) []gates.EvidenceRef {
	refs := []gates.EvidenceRef{
		{Kind: "task_payload", Path: task.ID, Note: "canonical task YAML in gate prompt"},
		{Kind: "spec", Path: feature, Note: "feature spec excerpt"},
	}
	if strings.TrimSpace(worktreePath) != "" {
		refs = append(refs, gates.EvidenceRef{
			Kind: "diff",
			Path: worktreePath,
			Note: "git diff excerpt from worktree",
		})
	}
	archPath := filepath.Join(repoRoot, "docs", "ai", "02-architecture.md")
	if _, err := os.Stat(archPath); err == nil {
		refs = append(refs, gates.EvidenceRef{
			Kind: "architecture_doc",
			Path: archPath,
			Note: "architecture excerpt when present",
		})
	}
	return refs
}

func (s *Service) buildGovernancePrompt(feature string, task sqlite.Task, worktreePath string) (string, error) {
	specExcerpt := ""
	if s.specReader != nil {
		if doc, err := s.specReader.ReadFeature(feature); err == nil && doc != nil {
			specExcerpt = truncateGovernanceText(doc.CombinedText(), governanceMaxExcerpt)
		}
	}
	canonical, _ := payloadToCanonical(task.PayloadJSON)
	taskYAML, err := yaml.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("marshal task for governance: %w", err)
	}
	diff := gitDiffForGovernance(worktreePath)
	archSection := ""
	if arch := readArchitectureExcerpt(s.repoRoot); arch != "" {
		archSection = "\n--- ARCHITECTURE ---\n" + arch
	}
	return fmt.Sprintf(governancePromptTemplate, specExcerpt, string(taskYAML), diff, archSection), nil
}

func readArchitectureExcerpt(repoRoot string) string {
	path := filepath.Join(repoRoot, "docs", "ai", "02-architecture.md")
	body, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return truncateGovernanceText(string(body), governanceMaxExcerpt)
}

func gitDiffForGovernance(worktreePath string) string {
	if strings.TrimSpace(worktreePath) == "" {
		return "(diff unavailable: empty worktree path)"
	}
	for _, args := range [][]string{
		{"diff", "HEAD"},
		{"diff"},
		{"diff", "--cached"},
	} {
		cmd := exec.Command("git", append([]string{"-C", worktreePath}, args...)...)
		out, err := cmd.Output()
		if err == nil {
			s := strings.TrimSpace(string(out))
			if s != "" {
				return truncateGovernanceText(s, governanceMaxExcerpt)
			}
		}
	}
	return "(no diff detected)"
}

func governanceRetries(task sqlite.Task) int {
	canonical, err := payloadToCanonical(task.PayloadJSON)
	if err != nil || canonical.Governance == nil {
		return 0
	}
	return canonical.Governance.Retries
}

func (s *Service) persistGovernanceVerdict(feature string, task sqlite.Task, v gates.Result, agentStdout string) error {
	at := time.Now().UTC().Format(time.RFC3339)
	retry := governanceRetries(task)
	entry := gateHistoryEntryFromResult(governanceGateName, v, at, retry)
	record := governanceRecordFromEntry(entry)

	canonical, err := payloadToCanonical(task.PayloadJSON)
	if err != nil {
		return err
	}
	if canonical.Governance == nil {
		canonical.Governance = &asagiri.TaskGovernance{}
	}
	canonical.Governance.History = append(canonical.Governance.History, record)
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
		task.ID, "task", governanceGateName, feature, s.cfg.GovernanceAgent(),
		"governance", "Governance gate", agentStdout, v,
	)
}

func (s *Service) incrementGovernanceRetries(task sqlite.Task) (int, sqlite.Task, error) {
	canonical, err := payloadToCanonical(task.PayloadJSON)
	if err != nil {
		return 0, task, err
	}
	if canonical.Governance == nil {
		canonical.Governance = &asagiri.TaskGovernance{}
	}
	canonical.Governance.Retries++
	retries := canonical.Governance.Retries
	canonical.TouchMetadata(time.Now().UTC())
	payload, err := canonicalToPayload(canonical)
	if err != nil {
		return 0, task, err
	}
	if err := s.store.UpdateTask(&sqlite.Task{ID: task.ID, PayloadJSON: payload}); err != nil {
		return 0, task, err
	}
	fresh, err := s.store.GetTask(task.ID)
	if err != nil {
		return retries, task, err
	}
	return retries, *fresh, nil
}

func (s *Service) maxGovernanceRetries() int {
	if s.cfg == nil {
		return 2
	}
	return s.cfg.Work.Gates.Governance.MaxRetriesValue()
}

func (s *Service) processGovernanceAfterDev(ctx context.Context, feature string, task sqlite.Task) (governanceOutcome, sqlite.Task, error) {
	if s.cfg == nil || !s.cfg.Work.Gates.Governance.IsActive() {
		return governanceOK, task, nil
	}

	worktreePath := strings.TrimSpace(task.WorktreePath)
	verdict, agentStdout, runErr := s.runGovernanceGate(ctx, feature, task, worktreePath)
	if runErr != nil {
		verdict = gates.Result{
			Status: gates.VerdictFail,
			Notes:  []string{runErr.Error()},
		}
	}
	if err := s.persistGovernanceVerdict(feature, task, verdict, agentStdout); err != nil {
		return governanceOK, task, err
	}
	if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
		task = *fresh
	}

	switch verdict.Status {
	case gates.VerdictPass:
		return governanceOK, task, nil
	case gates.VerdictWarn:
		if s.cfg.Work.Gates.Governance.WarnAdvisory() {
			return governanceOK, task, nil
		}
		if err := s.transitionTask(task, asagiri.StatusFailed, true); err != nil {
			return governanceOK, task, err
		}
		return governanceOK, task, fmt.Errorf("governance gate warn (non-advisory): %s", gates.FormatFailure(verdict))
	case gates.VerdictFail:
		used := governanceRetries(task)
		max := s.maxGovernanceRetries()
		if used < max {
			_, updated, err := s.incrementGovernanceRetries(task)
			if err != nil {
				return governanceOK, task, err
			}
			task = updated
			if err := s.transitionTask(task, asagiri.StatusRunning, true); err != nil {
				return governanceOK, task, err
			}
			if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
				task = *fresh
			}
			return governanceRetryDev, task, nil
		}
		if err := s.transitionTask(task, asagiri.StatusFailed, true); err != nil {
			return governanceOK, task, err
		}
		if fresh, getErr := s.store.GetTask(task.ID); getErr == nil {
			task = *fresh
		}
		return governanceOK, task, fmt.Errorf(
			"governance gate failed after %d retries (max %d): %s",
			used,
			max,
			gates.FormatFailure(verdict),
		)
	default:
		return governanceOK, task, fmt.Errorf("governance gate unknown status %q", verdict.Status)
	}
}

func (s *Service) applyGovernanceAfterDev(ctx context.Context, feature string, task sqlite.Task, worktreePath string) error {
	if worktreePath != "" && strings.TrimSpace(task.WorktreePath) == "" {
		task.WorktreePath = worktreePath
	}
	outcome, _, err := s.processGovernanceAfterDev(ctx, feature, task)
	if err != nil {
		return err
	}
	if outcome == governanceRetryDev {
		return fmt.Errorf("governance gate requested dev retry without retry loop")
	}
	return nil
}

func (s *Service) warnGovernanceInactiveMode(feature string) {
	if s.cfg == nil || !s.cfg.Work.Gates.Governance.EnabledButInactive() {
		return
	}
	msg := fmt.Sprintf(
		"governance: enabled=true but mode=%q is inactive (only per-task runs); gate skipped for feature %s",
		s.cfg.Work.Gates.Governance.Mode,
		feature,
	)
	_ = s.writeTaskLog("_config", "governance-warn.log", msg+"\n")
}
