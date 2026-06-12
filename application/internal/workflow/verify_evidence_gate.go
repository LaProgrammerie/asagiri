package workflow

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentresolve"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	appvalidate "github.com/LaProgrammerie/asagiri/application/internal/validation"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

const verifyEvidenceGatePromptTemplate = `Tu es un validateur read-only de preuves de validation post-verify. Tu ne produis AUCUN code et tu ne modifies aucun fichier.
Analyse si les résultats de validation locale sont suffisants et cohérents avec le payload de tâche (validation_commands, scope, actionnabilité).
Réponds UNIQUEMENT avec un bloc YAML de la forme:

verify_evidence_gate:
  status: pass|warn|fail
  confidence: 0.0-1.0
  notes:
    - ...
  findings:
    - code: missing_validation_commands|validation_output_missing|validation_output_empty|validation_evidence_incomplete|validation_scope_mismatch|weak_assertion_signal|missing_expected_command|verification_not_actionable|other
      severity: warn|fail
      message: ...
      actions:
        - ...

--- SPEC / REQUIREMENTS ---
%s

--- TASK PAYLOAD ---
%s

--- VALIDATION RESULTS ---
%s

--- VALIDATION LOG ---
%s
`

const verifyEvidenceGateName = "verify_evidence"

var verifyEvidenceGateParseConfig = gates.ParseConfig{
	BlockKey:          "verify_evidence_gate",
	MissingBlockError: "verify_evidence_gate block missing from agent output",
	ParseErrorNote:    "verify_evidence_gate_parse_error",
}

func (s *Service) processVerifyEvidenceGate(ctx context.Context, feature string, task sqlite.Task, results []appvalidate.Result) error {
	if s.cfg == nil || !s.cfg.Work.Gates.VerifyEvidence.IsActive() {
		return nil
	}

	result, agentStdout, runErr := s.runVerifyEvidenceGate(ctx, feature, task, results)
	if runErr != nil {
		result = gates.Result{
			GateID:   "verify_evidence_gate",
			GateType: "verify_evidence_gate",
			Scope:    task.ID,
			Status:   gates.VerdictFail,
			Notes:    []string{runErr.Error()},
		}
	}
	if err := s.persistVerifyEvidenceGateVerdict(feature, task, result, agentStdout); err != nil {
		return err
	}
	return gateOutcomeError("verify evidence gate", result, s.cfg.Work.Gates.VerifyEvidence.WarnAdvisory())
}

func (s *Service) runVerifyEvidenceGate(ctx context.Context, feature string, task sqlite.Task, results []appvalidate.Result) (gates.Result, string, error) {
	resultsPath := validationResultsPath(s.repoRoot, task.ID)
	evidence := verifyEvidenceGateEvidenceRefs(feature, task.ID, task.WorktreePath, task.PayloadJSON, results, resultsPath)
	if s.dryRun {
		return s.gateDryRunResult("verify_evidence_gate", "verify_evidence_gate", task.ID, "verify evidence gate dry-run: simulated pass", evidence), "", nil
	}

	legacyPrompt, err := s.buildVerifyEvidenceGatePrompt(feature, task.PayloadJSON, task.WorktreePath, results, resultsPath)
	if err != nil {
		return gates.Result{}, "", err
	}
	agentKey := s.cfg.VerifyEvidenceGateAgent()
	canonical, _ := payloadToCanonical(task.PayloadJSON)
	contextFiles := s.contextFilesForTask(feature, canonical)
	prompt, err := s.resolveGatePrompt(agentresolve.PhaseVerifyEvidence, agentKey, feature, task.ID, "", legacyPrompt, contextFiles)
	if err != nil {
		return gates.Result{}, "", err
	}

	workDir := s.repoRoot
	if strings.TrimSpace(task.WorktreePath) != "" {
		workDir = task.WorktreePath
	}

	stdout, err := s.executeGateAgent(ctx, agentKey, feature, task.ID, workDir, prompt, s.verifyEvidenceGateAgentHook)
	if err != nil {
		return gates.Result{}, stdout, err
	}

	parsed := gates.ParseResult(stdout, verifyEvidenceGateParseConfig)
	parsed.GateID = "verify_evidence_gate"
	parsed.GateType = "verify_evidence_gate"
	parsed.Scope = task.ID
	parsed.Evidence = evidence
	return gates.ClassifyResult(parsed, s.cfg.Work.Gates.VerifyEvidence.FailOn), stdout, nil
}

func (s *Service) buildVerifyEvidenceGatePrompt(feature, payloadJSON, worktreePath string, results []appvalidate.Result, resultsPath string) (string, error) {
	specExcerpt := ""
	if s.specReader != nil {
		if doc, err := s.specReader.ReadFeature(feature); err == nil && doc != nil {
			specExcerpt = truncateGovernanceText(doc.CombinedText(), governanceMaxExcerpt)
		}
	}
	payloadSection := strings.TrimSpace(payloadJSON)
	if payloadSection == "" {
		payloadSection = "(empty payload)"
	}
	resultsSection, err := formatValidationResultsForPrompt(results)
	if err != nil {
		return "", fmt.Errorf("format validation results for verify evidence gate: %w", err)
	}
	logSection := "(not persisted)"
	if strings.TrimSpace(resultsPath) != "" {
		if _, err := os.Stat(resultsPath); err == nil {
			logSection = resultsPath
		}
	}
	if strings.TrimSpace(worktreePath) != "" {
		logSection += fmt.Sprintf("\nworktree: %s", worktreePath)
	}
	return fmt.Sprintf(verifyEvidenceGatePromptTemplate, specExcerpt, payloadSection, resultsSection, logSection), nil
}

type validationResultPromptRow struct {
	Name     string `yaml:"name"`
	Command  string `yaml:"command"`
	ExitCode int    `yaml:"exit_code"`
	Output   string `yaml:"output,omitempty"`
}

func formatValidationResultsForPrompt(results []appvalidate.Result) (string, error) {
	if len(results) == 0 {
		return "(no validation results)", nil
	}
	perOutput := governanceMaxExcerpt / len(results)
	if perOutput < 512 {
		perOutput = 512
	}
	rows := make([]validationResultPromptRow, len(results))
	for i, r := range results {
		rows[i] = validationResultPromptRow{
			Name:     r.Name,
			Command:  r.Command,
			ExitCode: r.ExitCode,
			Output:   truncateGovernanceText(r.Output, perOutput),
		}
	}
	body, err := yaml.Marshal(rows)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func verifyEvidenceGateEvidenceRefs(feature, taskID, worktreePath, payloadJSON string, results []appvalidate.Result, resultsPath string) []gates.EvidenceRef {
	refs := []gates.EvidenceRef{
		{Kind: "task", Path: taskID, Note: "task verify evidence gate scope"},
		{Kind: "feature", Path: feature, Note: "feature scope"},
		{Kind: "payload", Path: taskID, Note: "task payload in prompt"},
	}
	if strings.TrimSpace(worktreePath) != "" {
		refs = append(refs, gates.EvidenceRef{
			Kind: "worktree",
			Path: worktreePath,
			Note: "validation worktree",
		})
	}
	if len(results) > 0 {
		refs = append(refs, gates.EvidenceRef{
			Kind: "validation",
			Path: taskID,
			Note: fmt.Sprintf("%d validation command result(s) in prompt", len(results)),
		})
	}
	if strings.TrimSpace(resultsPath) != "" {
		if _, err := os.Stat(resultsPath); err == nil {
			refs = append(refs, gates.EvidenceRef{
				Kind: "validation",
				Path: resultsPath,
				Note: "persisted validation/results.json",
			})
		}
	}
	if strings.TrimSpace(payloadJSON) != "" {
		refs = append(refs, gates.EvidenceRef{
			Kind: "spec",
			Path: fmt.Sprintf(".kiro/specs/%s/", feature),
			Note: "spec excerpt in prompt when available",
		})
	}
	return refs
}

func (s *Service) persistVerifyEvidenceGateVerdict(feature string, task sqlite.Task, v gates.Result, agentStdout string) error {
	at := time.Now().UTC().Format(time.RFC3339)
	entry := gateHistoryEntryFromResult(verifyEvidenceGateName, v, at, 0)

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
		task.ID, "task", verifyEvidenceGateName, feature, s.cfg.VerifyEvidenceGateAgent(),
		"verify_evidence_gate", "Verify evidence gate", agentStdout, v,
	)
}
