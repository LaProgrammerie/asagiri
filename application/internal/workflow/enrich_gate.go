package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentresolve"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

const enrichGatePromptTemplate = `Tu es un validateur de readiness dev read-only. Tu ne produis AUCUN code et tu ne modifies aucun fichier.
Le plan de tâches a déjà été validé : ne re-audite PAS la couverture macro des exigences.
Analyse si le payload enrichi est prêt pour l'implémentation (files_scope, validation_commands, contexte, actionnabilité).
Réponds UNIQUEMENT avec un bloc YAML de la forme:

enrich_gate:
  status: pass|warn|fail
  confidence: 0.0-1.0
  notes:
    - ...
  findings:
    - code: missing_files_scope|invalid_validation_commands|enrichment_not_actionable|empty_context_when_required|scope_too_broad|other
      severity: warn|fail
      message: ...
      actions:
        - ...

Règle : empty_context_when_required doit toujours être severity warn (jamais fail).

--- SPEC / REQUIREMENTS ---
%s

--- ENRICHED TASK PAYLOAD ---
%s

--- CONTEXT FILES ---
%s
`

const enrichGateName = "enrich"

var enrichGateParseConfig = gates.ParseConfig{
	BlockKey:          "enrich_gate",
	MissingBlockError: "enrich_gate block missing from agent output",
	ParseErrorNote:    "enrich_gate_parse_error",
}

func (s *Service) processEnrichGate(ctx context.Context, feature string, task sqlite.Task, payloadJSON string, contextFiles []string) error {
	if s.cfg == nil || !s.cfg.Work.Gates.Enrich.IsActive() {
		return nil
	}

	result, agentStdout, runErr := s.runEnrichGate(ctx, feature, task, payloadJSON, contextFiles)
	if runErr != nil {
		result = gates.Result{
			GateID:   "enrich_gate",
			GateType: "enrich_gate",
			Scope:    task.ID,
			Status:   gates.VerdictFail,
			Notes:    []string{runErr.Error()},
		}
	}
	if err := s.persistEnrichGateVerdict(feature, task, result, agentStdout); err != nil {
		return err
	}
	return gateOutcomeError("enrich gate", result, s.cfg.Work.Gates.Enrich.WarnAdvisory())
}

func (s *Service) runEnrichGate(ctx context.Context, feature string, task sqlite.Task, payloadJSON string, contextFiles []string) (gates.Result, string, error) {
	evidence := enrichGateEvidenceRefs(feature, task.ID, payloadJSON, contextFiles)
	if s.dryRun {
		return s.gateDryRunResult("enrich_gate", "enrich_gate", task.ID, "enrich gate dry-run: simulated pass", evidence), "", nil
	}

	legacyPrompt, err := s.buildEnrichGatePrompt(feature, payloadJSON, contextFiles)
	if err != nil {
		return gates.Result{}, "", err
	}
	agentKey := s.cfg.EnrichGateAgent()
	prompt, err := s.resolveGatePrompt(agentresolve.PhaseEnrichGate, agentKey, feature, task.ID, "", legacyPrompt, contextFiles)
	if err != nil {
		return gates.Result{}, "", err
	}

	stdout, err := s.executeGateAgent(ctx, agentKey, feature, task.ID, s.repoRoot, prompt, s.enrichGateAgentHook)
	if err != nil {
		return gates.Result{}, stdout, err
	}

	parsed := gates.ParseResult(stdout, enrichGateParseConfig)
	parsed.GateID = "enrich_gate"
	parsed.GateType = "enrich_gate"
	parsed.Scope = task.ID
	parsed.Evidence = evidence
	return gates.ClassifyResult(parsed, s.cfg.Work.Gates.Enrich.FailOn), stdout, nil
}

func (s *Service) buildEnrichGatePrompt(feature, payloadJSON string, contextFiles []string) (string, error) {
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
	contextSection := "(none)"
	if len(contextFiles) > 0 {
		contextYAML, err := yaml.Marshal(contextFiles)
		if err != nil {
			return "", fmt.Errorf("marshal context files for enrich gate: %w", err)
		}
		contextSection = string(contextYAML)
	}
	return fmt.Sprintf(enrichGatePromptTemplate, specExcerpt, payloadSection, contextSection), nil
}

func enrichGateEvidenceRefs(feature, taskID, payloadJSON string, contextFiles []string) []gates.EvidenceRef {
	refs := []gates.EvidenceRef{
		{Kind: "task", Path: taskID, Note: "task enrich gate scope"},
		{Kind: "feature", Path: feature, Note: "feature scope"},
		{Kind: "payload", Path: taskID, Note: "enriched payload candidate in prompt"},
	}
	if strings.TrimSpace(payloadJSON) != "" {
		refs = append(refs, gates.EvidenceRef{
			Kind: "spec",
			Path: fmt.Sprintf(".kiro/specs/%s/", feature),
			Note: "spec excerpt in prompt when available",
		})
	}
	if len(contextFiles) > 0 {
		refs = append(refs, gates.EvidenceRef{
			Kind: "context",
			Path: taskID,
			Note: fmt.Sprintf("%d context file(s) in prompt", len(contextFiles)),
		})
	}
	return refs
}

func (s *Service) persistEnrichGateVerdict(feature string, task sqlite.Task, v gates.Result, agentStdout string) error {
	at := time.Now().UTC().Format(time.RFC3339)
	entry := gateHistoryEntryFromResult(enrichGateName, v, at, 0)

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
		task.ID, "task", enrichGateName, feature, s.cfg.EnrichGateAgent(),
		"enrich_gate", "Enrich gate", agentStdout, v,
	)
}

func mergeEnrichPayloadAfterGate(existingPayload string, enriched map[string]any) (string, error) {
	base := map[string]any{}
	if strings.TrimSpace(existingPayload) != "" {
		if err := json.Unmarshal([]byte(existingPayload), &base); err != nil {
			return "", fmt.Errorf("merge enrich payload: %w", err)
		}
	}
	for k, v := range enriched {
		base[k] = v
	}
	body, err := json.Marshal(base)
	if err != nil {
		return "", fmt.Errorf("marshal merged enrich payload: %w", err)
	}
	return string(body), nil
}
