package worktrust

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
)

// WorkTrustReportToGateResult maps a synthesized trust report to a work gate result without changing scoring.
func WorkTrustReportToGateResult(report WorkTrustReport, cfg config.WorkTrustGateConfig) gates.Result {
	scope := report.Scope.TaskID
	if scope == "" {
		scope = report.Scope.ID
	}
	result := gates.Result{
		GateID:     "trust_gate",
		GateType:   "trust_gate",
		Scope:      scope,
		Confidence: report.Score.Overall / 100,
		Notes:      trustGateNotes(report),
		Findings:   trustGateFindings(report),
		Evidence:   trustGateEvidence(report),
	}
	if result.Confidence < 0 {
		result.Confidence = 0
	}
	if result.Confidence > 1 {
		result.Confidence = 1
	}

	verdict := strings.ToLower(string(report.Score.Verdict))
	if trustVerdictListed(verdict, cfg.BlockVerdicts) || report.Score.Overall < cfg.MinScoreValue() {
		result.Status = gates.VerdictFail
		if report.Score.Overall < cfg.MinScoreValue() {
			result.Findings = append(result.Findings, gates.Finding{
				Code:     "trust_score_below_min",
				Severity: "fail",
				Message:  fmt.Sprintf("trust score %.1f below minimum %.1f", report.Score.Overall, cfg.MinScoreValue()),
				Actions:  []string{fmt.Sprintf("asa trust task %s", scope)},
			})
		}
		return result
	}
	if trustVerdictListed(verdict, cfg.WarnVerdicts) {
		result.Status = gates.VerdictWarn
		return result
	}
	result.Status = gates.VerdictPass
	return result
}

func trustVerdictListed(verdict string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), verdict) {
			return true
		}
	}
	return false
}

func trustGateNotes(report WorkTrustReport) []string {
	if s := strings.TrimSpace(report.Score.Summary); s != "" {
		return []string{s}
	}
	return []string{fmt.Sprintf("trust verdict %s score %.1f", report.Score.Verdict, report.Score.Overall)}
}

func trustGateFindings(report WorkTrustReport) []gates.Finding {
	if len(report.Findings) == 0 {
		return nil
	}
	out := make([]gates.Finding, 0, len(report.Findings))
	for _, f := range report.Findings {
		severity := "warn"
		if strings.EqualFold(f.Severity, "fail") || strings.EqualFold(f.Severity, "high") {
			severity = "fail"
		}
		out = append(out, gates.Finding{
			Code:     f.Code,
			Severity: severity,
			Message:  f.Message,
			Actions:  append([]string(nil), f.Actions...),
		})
	}
	return out
}

func trustGateEvidence(report WorkTrustReport) []gates.EvidenceRef {
	if len(report.Evidences) == 0 {
		return nil
	}
	out := make([]gates.EvidenceRef, 0, len(report.Evidences))
	for _, e := range report.Evidences {
		out = append(out, gates.EvidenceRef{
			Kind: e.Kind,
			Path: e.Ref,
			Note: e.Summary,
		})
	}
	return out
}
