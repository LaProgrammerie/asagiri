package worktrust

import (
	"fmt"
	"strings"
)

// FormatDailyNextBlock is a compact trust summary for `asa next`.
func FormatDailyNextBlock(report WorkTrustReport) string {
	var sb strings.Builder
	writeSectionHeader(&sb, "Trust")
	fmt.Fprintf(&sb, "  Verdict: %s\n", VerdictLabel(report.Score.Verdict))
	if risk := mainRiskMessage(report); risk != "" {
		fmt.Fprintf(&sb, "  Risque: %s\n", risk)
	}
	rec := report.Recommendation
	if rec.Command != "" {
		fmt.Fprintf(&sb, "  → %s\n", rec.Command)
	} else if rec.Rationale != "" {
		fmt.Fprintf(&sb, "  %s\n", rec.Rationale)
	}
	return sb.String()
}

// FormatDailyStatusBlock is a compact trust section for `asa status`.
func FormatDailyStatusBlock(report FeatureTrustReport) string {
	var sb strings.Builder
	writeSectionHeader(&sb, "Trust")
	atRisk := tasksAtRisk(report.Tasks)
	riskyLabel := "aucune"
	if n := len(atRisk); n > 0 {
		if n == 1 {
			riskyLabel = "1 task à risque"
		} else {
			riskyLabel = fmt.Sprintf("%d tasks à risque", n)
		}
	}
	fmt.Fprintf(&sb, "  Feature: %s  Verdict: %s  %s\n",
		report.Scope.Feature, VerdictLabel(report.Score.Verdict), riskyLabel)
	if len(report.NextActions) > 0 && report.NextActions[0].Command != "" {
		fmt.Fprintf(&sb, "  → %s\n", report.NextActions[0].Command)
	}
	return sb.String()
}

// FormatDailyPostWorkLine is a single-line trust summary after work/continue.
func FormatDailyPostWorkLine(scope string, report WorkTrustReport) string {
	scope = strings.TrimSpace(scope)
	label := VerdictLabel(report.Score.Verdict)
	parts := []string{fmt.Sprintf("Trust: %s", label)}
	if scope != "" {
		parts[0] = fmt.Sprintf("Trust: %s (%s)", label, scope)
	}
	if risk := mainRiskMessage(report); risk != "" {
		parts = append(parts, risk)
	}
	if id := strings.TrimSpace(report.Scope.TaskID); id != "" {
		parts = append(parts, fmt.Sprintf("asa trust task %s", id))
	} else if id := strings.TrimSpace(report.Scope.ID); id != "" && report.Scope.Kind == "run" {
		parts = append(parts, fmt.Sprintf("asa trust run %s", id))
	}
	return strings.Join(parts, " — ")
}

func mainRiskMessage(report WorkTrustReport) string {
	if risks := topFindings(report.Findings, 1); len(risks) > 0 {
		f := risks[0]
		src := strings.TrimSpace(f.Source)
		msg := strings.TrimSpace(f.Message)
		if src != "" && msg != "" {
			return src + " — " + msg
		}
		if msg != "" {
			return msg
		}
		return src
	}
	for _, d := range report.Dimensions {
		if d.Status == DimStatusFailed || d.Status == DimStatusWeak {
			if d.Summary != "" {
				return d.Label + " — " + d.Summary
			}
		}
	}
	if report.Score.Verdict == VerdictTrusted || report.Score.Verdict == VerdictAcceptable {
		return ""
	}
	return verdictNarrative(report.Score.Verdict, report.Score.Summary)
}

// FormatDailyPostWorkFromRun picks the worst task or falls back to run scope.
func FormatDailyPostWorkFromRun(report RunTrustReport) string {
	for _, t := range report.Tasks {
		if verdictRank(t.Verdict) >= verdictRank(VerdictRisky) {
			scope := t.TaskID
			if report.Scope.Feature != "" {
				scope = report.Scope.Feature + " / " + t.TaskID
			}
			lineReport := WorkTrustReport{
				Scope: TrustScope{Kind: "task", ID: t.TaskID, Feature: report.Scope.Feature, TaskID: t.TaskID, Status: t.Status},
				Score: WorkTrustScore{Verdict: t.Verdict, Overall: t.Score},
			}
			if risk := mainRiskForTaskSummary(t); risk != "" {
				lineReport.Findings = []WorkTrustFinding{{Source: t.TaskID, Message: risk}}
			}
			return FormatDailyPostWorkLine(scope, lineReport)
		}
	}
	scope := report.Scope.ID
	if report.Scope.Feature != "" {
		scope = report.Scope.Feature + " / " + report.Scope.ID
	}
	lineReport := WorkTrustReport{
		Scope: report.Scope,
		Score: report.Score,
	}
	return FormatDailyPostWorkLine(scope, lineReport)
}

func mainRiskForTaskSummary(t FeatureTaskSummary) string {
	if verdictRank(t.Verdict) >= verdictRank(VerdictBlocked) {
		return string(t.Verdict)
	}
	return ""
}
