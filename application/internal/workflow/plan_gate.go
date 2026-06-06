package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/spec"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

const planGatePromptTemplate = `Tu es un validateur de plan read-only. Tu ne produis AUCUN code et tu ne modifies aucun fichier.
Analyse si le plan de tâches couvre les exigences, est bien découpé et cohérent.
Réponds UNIQUEMENT avec un bloc YAML de la forme:

plan_gate:
  status: pass|warn|fail
  confidence: 0.0-1.0
  notes:
    - ...
  findings:
    - code: missing_requirement_coverage|oversized_task|invalid_dependency|other
      severity: warn|fail
      message: ...
      actions:
        - ...

--- SPEC / REQUIREMENTS ---
%s

--- PLANNED TASKS ---
%s
`

const planGateName = "plan"

var planGateParseConfig = gates.ParseConfig{
	BlockKey:          "plan_gate",
	MissingBlockError: "plan_gate block missing from agent output",
	ParseErrorNote:    "plan_gate_parse_error",
}

func (s *Service) processPlanGate(ctx context.Context, run *sqlite.Run, feature string, doc *spec.Document, canonicalTasks []asagiri.Task) error {
	if s.cfg == nil || !s.cfg.Work.Gates.Plan.IsActive() {
		return nil
	}

	result, agentStdout, runErr := s.runPlanGate(ctx, feature, doc, canonicalTasks)
	if runErr != nil {
		result = gates.Result{
			GateID:   "plan_gate",
			GateType: "plan_gate",
			Scope:    feature,
			Status:   gates.VerdictFail,
			Notes:    []string{runErr.Error()},
		}
	}
	if err := s.persistPlanGateResult(run, feature, result, agentStdout); err != nil {
		return err
	}
	return gateOutcomeError("plan gate", result, s.cfg.Work.Gates.Plan.WarnAdvisory())
}

func (s *Service) runPlanGate(ctx context.Context, feature string, doc *spec.Document, canonicalTasks []asagiri.Task) (gates.Result, string, error) {
	evidence := planGateEvidenceRefs(feature, doc, canonicalTasks)
	if s.dryRun {
		return s.gateDryRunResult("plan_gate", "plan_gate", feature, "plan gate dry-run: simulated pass", evidence), "", nil
	}

	prompt, err := buildPlanGatePrompt(doc, canonicalTasks)
	if err != nil {
		return gates.Result{}, "", err
	}

	stdout, err := s.executeGateAgent(ctx, s.cfg.PlanGateAgent(), feature, "", s.repoRoot, prompt, s.planGateAgentHook)
	if err != nil {
		return gates.Result{}, stdout, err
	}

	parsed := gates.ParseResult(stdout, planGateParseConfig)
	parsed.GateID = "plan_gate"
	parsed.GateType = "plan_gate"
	parsed.Scope = feature
	parsed.Evidence = evidence
	return gates.ClassifyResult(parsed, s.cfg.Work.Gates.Plan.FailOn), stdout, nil
}

func buildPlanGatePrompt(doc *spec.Document, canonicalTasks []asagiri.Task) (string, error) {
	specExcerpt := ""
	if doc != nil {
		specExcerpt = truncateGovernanceText(doc.CombinedText(), governanceMaxExcerpt)
	}
	tasksYAML, err := yaml.Marshal(canonicalTasks)
	if err != nil {
		return "", fmt.Errorf("marshal planned tasks for plan gate: %w", err)
	}
	return fmt.Sprintf(planGatePromptTemplate, specExcerpt, string(tasksYAML)), nil
}

func planGateEvidenceRefs(feature string, doc *spec.Document, canonicalTasks []asagiri.Task) []gates.EvidenceRef {
	refs := []gates.EvidenceRef{
		{Kind: "feature", Path: feature, Note: "planned feature scope"},
		{Kind: "planned_tasks", Path: feature, Note: fmt.Sprintf("%d tasks in plan gate prompt", len(canonicalTasks))},
	}
	if doc != nil {
		if strings.TrimSpace(doc.Requirements) != "" {
			refs = append(refs, gates.EvidenceRef{
				Kind: "requirements",
				Path: fmt.Sprintf(".kiro/specs/%s/requirements.md", feature),
				Note: "requirements excerpt in prompt",
			})
		}
		if strings.TrimSpace(doc.Tasks) != "" {
			refs = append(refs, gates.EvidenceRef{
				Kind: "spec_tasks",
				Path: fmt.Sprintf(".kiro/specs/%s/tasks.md", feature),
				Note: "source tasks.md",
			})
		}
	}
	return refs
}

func (s *Service) persistPlanGateResult(run *sqlite.Run, feature string, r gates.Result, agentStdout string) error {
	return s.persistGateLogs(
		run.ID, "run", planGateName, feature, s.cfg.PlanGateAgent(),
		"plan_gate", "Plan gate", agentStdout, r,
	)
}
