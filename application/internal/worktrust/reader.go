package worktrust

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

// validationEvidenceDocument mirrors workflow persistence (read-only, no import cycle).
type validationEvidenceDocument struct {
	TaskID   string                      `json:"task_id"`
	Worktree string                      `json:"worktree,omitempty"`
	DryRun   bool                        `json:"dry_run,omitempty"`
	At       string                      `json:"at"`
	Commands []validationEvidenceCommand `json:"commands"`
}

type validationEvidenceCommand struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output,omitempty"`
}

type taskSignals struct {
	RepoRoot   string
	Task       sqlite.Task
	Cfg        *config.Config
	Entries    map[string]asagiri.GateHistoryEntry
	Validation *validationEvidenceDocument
	Pending    gates.PendingGate
	HasPending bool
	EnrichOK   bool
	VerifyEvOK bool
	Findings   []WorkTrustFinding
	Evidences  []WorkTrustEvidence
}

// BuildTaskReport synthesizes a read-only trust report for one task.
func BuildTaskReport(repoRoot string, cfg *config.Config, task sqlite.Task) (WorkTrustReport, error) {
	signals, err := collectTaskSignals(repoRoot, cfg, task)
	if err != nil {
		return WorkTrustReport{}, err
	}
	scored := scoreTask(signals)
	scored.Recommendation = buildRecommendation(signals, scored)
	return scored, nil
}

func collectTaskSignals(repoRoot string, cfg *config.Config, task sqlite.Task) (taskSignals, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	s := taskSignals{
		RepoRoot:   strings.TrimSpace(repoRoot),
		Task:       task,
		Cfg:        cfg,
		Entries:    lastGateEntriesByName(task.PayloadJSON),
		EnrichOK:   gates.EnrichGateSatisfied(cfg, task.PayloadJSON),
		VerifyEvOK: gates.VerifyEvidenceGateSatisfied(cfg, task.PayloadJSON),
	}
	if pg, ok := gates.BlockingPendingForTask(s.RepoRoot, cfg, task); ok {
		s.Pending = pg
		s.HasPending = true
	}
	if doc, ok := readValidationEvidence(s.RepoRoot, task.ID); ok {
		s.Validation = &doc
		s.Evidences = append(s.Evidences, WorkTrustEvidence{
			Kind:    "validation",
			Ref:     validationResultsPath(s.RepoRoot, task.ID),
			Summary: fmt.Sprintf("%d command(s)", len(doc.Commands)),
			At:      doc.At,
		})
	}
	s.mergeGateLogFallbacks()
	s.collectFindingsAndEvidences()
	return s, nil
}

func lastGateEntriesByName(payloadJSON string) map[string]asagiri.GateHistoryEntry {
	out := make(map[string]asagiri.GateHistoryEntry)
	if payloadJSON == "" {
		return out
	}
	var payload asagiri.Task
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return out
	}
	if payload.Gates != nil {
		for _, e := range payload.Gates.History {
			name := normalizeGateName(e.Gate)
			out[name] = e
		}
	}
	if payload.Governance != nil && len(payload.Governance.History) > 0 {
		if _, ok := out["governance"]; !ok {
			last := payload.Governance.History[len(payload.Governance.History)-1]
			out["governance"] = asagiri.GateHistoryEntry{
				Gate:       "governance",
				At:         last.At,
				Status:     last.Status,
				Confidence: last.Confidence,
				Notes:      last.Notes,
				Findings:   last.Findings,
				Retry:      last.Retry,
				DryRun:     last.DryRun,
				ParseError: last.ParseError,
			}
		}
	}
	return out
}

func normalizeGateName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return "governance"
	}
	return n
}

func (s *taskSignals) mergeGateLogFallbacks() {
	if s.RepoRoot == "" {
		return
	}
	for _, gateName := range []string{"enrich", "governance", "human_review", "verify_evidence", "plan"} {
		if _, ok := s.Entries[gateName]; ok {
			continue
		}
		if !s.gateActive(gateName) && gateName != "plan" {
			continue
		}
		row, ok := gateRowFromLogFile(s.RepoRoot, s.Task.ID, gateName)
		if !ok {
			continue
		}
		s.Entries[gateName] = asagiri.GateHistoryEntry{
			Gate:       gateName,
			Status:     row.Status,
			Confidence: row.Confidence,
			Notes:      row.Notes,
		}
	}
}

type gateLogRow struct {
	Status     string
	Confidence float64
	Notes      []string
}

func gateRowFromLogFile(repoRoot, scopeID, gateName string) (gateLogRow, bool) {
	path := filepath.Join(repoRoot, ".asagiri", "logs", scopeID, "gates", gateName+".json")
	body, err := os.ReadFile(path)
	if err != nil {
		return gateLogRow{}, false
	}
	var doc gates.LogDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return gateLogRow{}, false
	}
	return gateLogRow{
		Status:     doc.Status,
		Confidence: doc.Confidence,
		Notes:      doc.Notes,
	}, true
}

func (s *taskSignals) gateActive(gateName string) bool {
	if s.Cfg == nil {
		return false
	}
	g := s.Cfg.Work.Gates
	switch gateName {
	case gates.EnrichGateName:
		return g.Enrich.IsActive()
	case "governance":
		return g.Governance.IsActive()
	case gates.HumanReviewGateName:
		return g.HumanReview.IsActive()
	case gates.VerifyEvidenceGateName:
		return g.VerifyEvidence.IsActive()
	case "plan":
		return g.Plan.IsActive()
	default:
		return false
	}
}

func (s *taskSignals) warnAdvisory(gateName string) bool {
	if s.Cfg == nil {
		return true
	}
	g := s.Cfg.Work.Gates
	switch gateName {
	case gates.EnrichGateName:
		return g.Enrich.WarnAdvisory()
	case "governance":
		return g.Governance.WarnAdvisory()
	case gates.HumanReviewGateName:
		return g.HumanReview.WarnAdvisory()
	case gates.VerifyEvidenceGateName:
		return g.VerifyEvidence.WarnAdvisory()
	case "plan":
		return g.Plan.WarnAdvisory()
	default:
		return true
	}
}

func (s *taskSignals) collectFindingsAndEvidences() {
	for gateName, entry := range s.Entries {
		s.Evidences = append(s.Evidences, WorkTrustEvidence{
			Kind:    "gate_history",
			Ref:     gateName,
			Summary: entry.Status,
			At:      entry.At,
		})
		for _, f := range entry.Findings {
			s.Findings = append(s.Findings, WorkTrustFinding{
				Code:     f.Code,
				Severity: mapFindingSeverity(f.Severity, entry.Status),
				Message:  f.Message,
				Source:   gateName,
				Actions:  f.Actions,
			})
		}
		st := strings.ToLower(strings.TrimSpace(entry.Status))
		if st == string(gates.VerdictFail) {
			s.Findings = append(s.Findings, WorkTrustFinding{
				Code:     gateName + "_fail",
				Severity: "high",
				Message:  fmt.Sprintf("gate %s failed", gateName),
				Source:   gateName,
			})
		} else if st == string(gates.VerdictWarn) && !gates.GateEntrySatisfied(s.warnAdvisory(gateName), entry) {
			s.Findings = append(s.Findings, WorkTrustFinding{
				Code:     gateName + "_warn",
				Severity: "medium",
				Message:  fmt.Sprintf("gate %s warn (non-advisory)", gateName),
				Source:   gateName,
			})
		}
	}
	if s.HasPending {
		s.Findings = append(s.Findings, WorkTrustFinding{
			Code:     "human_review_pending",
			Severity: "critical",
			Message:  "human review gate requires operator action",
			Source:   gates.HumanReviewGateName,
		})
	}
}

func validationResultsPath(repoRoot, taskID string) string {
	return filepath.Join(repoRoot, ".asagiri", "logs", taskID, "validation", "results.json")
}

func readValidationEvidence(repoRoot, taskID string) (validationEvidenceDocument, bool) {
	if repoRoot == "" || taskID == "" {
		return validationEvidenceDocument{}, false
	}
	body, err := os.ReadFile(validationResultsPath(repoRoot, taskID))
	if err != nil {
		return validationEvidenceDocument{}, false
	}
	var doc validationEvidenceDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return validationEvidenceDocument{}, false
	}
	return doc, true
}

func buildRecommendation(s taskSignals, report WorkTrustReport) WorkTrustRecommendation {
	priority := "suggested"
	if report.Score.Verdict == VerdictBlocked || report.Score.Verdict == VerdictRisky {
		priority = "required"
	}
	return WorkTrustRecommendation{
		Action:    "next",
		Command:   NextCommandForFeature(s.Task.Feature),
		Rationale: "continuer le workflow",
		Priority:  priority,
	}
}

func scoreTask(s taskSignals) WorkTrustReport {
	dims := scoreDimensions(s)
	findings := append([]WorkTrustFinding(nil), s.Findings...)
	overall := computeOverall(dims)
	overall = applyStatusCaps(s.Task.Status, overall)
	verdict := computeVerdict(s, overall, dims)
	if s.HasPending && overall > 40 {
		overall = 40
	}

	report := WorkTrustReport{
		ReportVersion: ReportVersion,
		Scope: TrustScope{
			Kind:    "task",
			ID:      s.Task.ID,
			Feature: s.Task.Feature,
			TaskID:  s.Task.ID,
			Status:  s.Task.Status,
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Score: WorkTrustScore{
			Overall: overall,
			Verdict: verdict,
			Summary: verdictSummary(verdict, overall),
		},
		Dimensions: dims,
		Findings:   findings,
		Evidences:  append([]WorkTrustEvidence(nil), s.Evidences...),
	}
	return report
}
