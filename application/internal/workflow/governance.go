package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
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

// governanceAgentHook replaces the governance agent Run in tests when set.
type governanceAgentHook func(ctx context.Context, prompt string) (stdout string, err error)

func (s *Service) runGovernanceGate(ctx context.Context, feature string, task sqlite.Task, worktreePath string) (governanceVerdict, string, error) {
	if s.dryRun {
		return governanceVerdict{
			Status:     "pass",
			Confidence: 1,
			Notes:      []string{"governance dry-run: simulated pass"},
			DryRun:     true,
		}, "", nil
	}

	prompt, err := s.buildGovernancePrompt(feature, task, worktreePath)
	if err != nil {
		return governanceVerdict{}, "", err
	}

	var stdout string
	if s.governanceAgentHook != nil {
		stdout, err = s.governanceAgentHook(ctx, prompt)
		if err != nil {
			return governanceVerdict{}, stdout, fmt.Errorf("governance agent: %w", err)
		}
	} else {
		agentName := s.cfg.GovernanceAgent()
		a, err := s.ensureAgent(agentName)
		if err != nil {
			return governanceVerdict{}, "", err
		}
		res, runErr := a.Run(ctx, agent.RunRequest{
			Feature:    feature,
			TaskID:     task.ID,
			Prompt:     prompt,
			WorkingDir: worktreePath,
		})
		if runErr != nil {
			return governanceVerdict{}, res.Stdout, fmt.Errorf("governance agent run: %w", runErr)
		}
		stdout = res.Stdout
		if stdout == "" {
			stdout = res.Stderr
		}
	}

	parsed := parseGovernanceVerdict(stdout)
	parsed.Status = classifyGovernanceVerdict(parsed, s.cfg.Work.Governance.FailOn)
	return parsed, stdout, nil
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

func (s *Service) persistGovernanceVerdict(feature string, task sqlite.Task, v governanceVerdict, agentStdout string) error {
	at := time.Now().UTC().Format(time.RFC3339)
	retry := governanceRetries(task)
	record := asagiri.GovernanceRecord{
		At:         at,
		Status:     v.Status,
		Confidence: v.Confidence,
		Notes:      v.Notes,
		Findings:   v.Findings,
		Retry:      retry,
		DryRun:     v.DryRun,
		ParseError: v.ParseError,
	}

	canonical, err := payloadToCanonical(task.PayloadJSON)
	if err != nil {
		return err
	}
	if canonical.Governance == nil {
		canonical.Governance = &asagiri.TaskGovernance{}
	}
	canonical.Governance.History = append(canonical.Governance.History, record)
	canonical.TouchMetadata(time.Now().UTC())

	payload, err := canonicalToPayload(canonical)
	if err != nil {
		return err
	}
	if err := s.store.UpdateTask(&sqlite.Task{ID: task.ID, PayloadJSON: payload}); err != nil {
		return err
	}

	doc := governanceLogDocument{
		TaskID:     task.ID,
		Feature:    feature,
		At:         at,
		Status:     v.Status,
		Confidence: v.Confidence,
		Notes:      v.Notes,
		Findings:   v.Findings,
		DryRun:     v.DryRun,
		ParseError: v.ParseError,
		Agent:      s.cfg.GovernanceAgent(),
	}
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := s.writeTaskLog(task.ID, "governance.json", string(body)+"\n"); err != nil {
		return err
	}
	return s.writeGovernanceAgentLog(task.ID, agentStdout, v)
}

func (s *Service) writeGovernanceAgentLog(taskID, agentStdout string, v governanceVerdict) error {
	var sb strings.Builder
	sb.WriteString("# Governance gate\n\n")
	sb.WriteString("## Agent stdout\n\n")
	if strings.TrimSpace(agentStdout) == "" {
		sb.WriteString("(empty)\n\n")
	} else {
		sb.WriteString(agentStdout)
		if !strings.HasSuffix(agentStdout, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("## Verdict\n\n")
	verdictYAML, err := yaml.Marshal(map[string]any{
		"governance": map[string]any{
			"status":      v.Status,
			"confidence":  v.Confidence,
			"notes":       v.Notes,
			"findings":    v.Findings,
			"dry_run":     v.DryRun,
			"parse_error": v.ParseError,
		},
	})
	if err != nil {
		sb.WriteString(fmt.Sprintf("status: %s\n", v.Status))
	} else {
		sb.Write(verdictYAML)
	}
	return s.writeTaskLog(taskID, "governance.log", sb.String())
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
	return s.cfg.Work.Governance.MaxRetriesValue()
}

func (s *Service) processGovernanceAfterDev(ctx context.Context, feature string, task sqlite.Task) (governanceOutcome, sqlite.Task, error) {
	if s.cfg == nil || !s.cfg.Work.Governance.IsActive() {
		return governanceOK, task, nil
	}

	worktreePath := strings.TrimSpace(task.WorktreePath)
	verdict, agentStdout, runErr := s.runGovernanceGate(ctx, feature, task, worktreePath)
	if runErr != nil {
		verdict = governanceVerdict{
			Status: "fail",
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
	case "pass":
		return governanceOK, task, nil
	case "warn":
		if s.cfg.Work.Governance.WarnAdvisory() {
			return governanceOK, task, nil
		}
		if err := s.transitionTask(task, asagiri.StatusFailed, true); err != nil {
			return governanceOK, task, err
		}
		return governanceOK, task, fmt.Errorf("governance gate warn (non-advisory): %s", formatGovernanceFailure(verdict))
	case "fail":
		// retries_used = relances déjà consommées ; max_retries = relances autorisées après le 1er FAIL.
		// Boucle : used < max → consommer une relance et reboucler dev ; sinon failed (anti-boucle : max fini).
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
			formatGovernanceFailure(verdict),
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
	if s.cfg == nil || !s.cfg.Work.Governance.EnabledButInactive() {
		return
	}
	msg := fmt.Sprintf(
		"governance: enabled=true but mode=%q is inactive (only per-task runs); gate skipped for feature %s",
		s.cfg.Work.Governance.Mode,
		feature,
	)
	_ = s.writeTaskLog("_config", "governance-warn.log", msg+"\n")
}
