package worktrust

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
)

const sectionRule = "─────"

// FormatTaskReport renders a human-readable trust summary for the terminal.
func FormatTaskReport(report WorkTrustReport, opts FormatOptions) string {
	var sb strings.Builder
	writeSectionHeader(&sb, "Summary")
	fmt.Fprintf(&sb, "  Verdict: %s\n", VerdictLabel(report.Score.Verdict))
	if opts.Explain {
		fmt.Fprintf(&sb, "  Score: %s\n", scoreLabel(report.Score.Overall))
	}
	if report.Score.Summary != "" && !opts.Explain {
		fmt.Fprintf(&sb, "  %s\n", verdictNarrative(report.Score.Verdict, report.Score.Summary))
	}
	if report.Scope.Feature != "" {
		fmt.Fprintf(&sb, "  Feature: %s  Task: %s  Status: %s\n",
			report.Scope.Feature, report.Scope.TaskID, HumanStatus(report.Scope.Status))
	} else {
		fmt.Fprintf(&sb, "  Task: %s  Status: %s\n", report.Scope.TaskID, HumanStatus(report.Scope.Status))
	}

	writeSectionHeader(&sb, "Gates")
	gateLines := taskGateLines(report)
	if len(gateLines) == 0 {
		sb.WriteString("  (aucune gate active)\n")
	} else {
		for _, line := range gateLines {
			sb.WriteString("  " + line + "\n")
		}
	}

	writeSectionHeader(&sb, "Risks")
	risks := topFindings(report.Findings, 5)
	if len(risks) == 0 {
		sb.WriteString("  (aucun)\n")
	} else {
		for _, f := range risks {
			fmt.Fprintf(&sb, "  • [%s] %s: %s\n", humanSeverity(f.Severity), f.Source, f.Message)
		}
	}

	writeSectionHeader(&sb, "Next actions")
	writeRecommendations(&sb, []WorkTrustRecommendation{report.Recommendation})

	if opts.Explain {
		writeSectionHeader(&sb, "Dimensions")
		for _, d := range report.Dimensions {
			fmt.Fprintf(&sb, "  %-22s %12s", d.Label, dimensionStatusLabel(d.Status))
			if d.Score >= 0 {
				fmt.Fprintf(&sb, "  %3.0f/100", d.Score)
			} else {
				sb.WriteString("       n/a")
			}
			if d.Summary != "" {
				fmt.Fprintf(&sb, "  %s", d.Summary)
			}
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// FormatFeatureReport renders a feature-level trust summary for the terminal.
func FormatFeatureReport(report FeatureTrustReport, opts FormatOptions) string {
	var sb strings.Builder
	writeSectionHeader(&sb, "Summary")
	fmt.Fprintf(&sb, "  Verdict: %s\n", VerdictLabel(report.Score.Verdict))
	if opts.Explain {
		fmt.Fprintf(&sb, "  Score: %s\n", scoreLabel(report.Score.Overall))
	} else {
		fmt.Fprintf(&sb, "  %s\n", verdictNarrative(report.Score.Verdict, report.Score.Summary))
	}
	fmt.Fprintf(&sb, "  Feature: %s  Tasks: %d\n", report.Scope.Feature, report.TaskCount)

	writeSectionHeader(&sb, "Gates")
	atRisk := tasksAtRisk(report.Tasks)
	if len(atRisk) == 0 {
		sb.WriteString("  Toutes les tasks sont dans une zone acceptable\n")
	} else {
		for _, t := range atRisk {
			if opts.Explain {
				fmt.Fprintf(&sb, "  %s  %s  %s  %s\n", t.TaskID, VerdictLabel(t.Verdict), scoreLabel(t.Score), HumanStatus(t.Status))
			} else {
				fmt.Fprintf(&sb, "  %s  %s  (%s)\n", t.TaskID, VerdictLabel(t.Verdict), HumanStatus(t.Status))
			}
		}
	}

	writeSectionHeader(&sb, "Risks")
	if len(atRisk) == 0 {
		sb.WriteString("  (aucun)\n")
	} else {
		fmt.Fprintf(&sb, "  %d task(s) bloquée(s) ou à surveiller\n", len(atRisk))
	}

	writeSectionHeader(&sb, "Next actions")
	writeRecommendations(&sb, report.NextActions)

	return sb.String()
}

// FormatRunReport renders a run-level trust summary for the terminal.
func FormatRunReport(report RunTrustReport, opts FormatOptions) string {
	var sb strings.Builder
	writeSectionHeader(&sb, "Summary")
	fmt.Fprintf(&sb, "  Verdict: %s\n", VerdictLabel(report.Score.Verdict))
	if opts.Explain {
		fmt.Fprintf(&sb, "  Score: %s\n", scoreLabel(report.Score.Overall))
	} else {
		fmt.Fprintf(&sb, "  %s\n", verdictNarrative(report.Score.Verdict, report.Score.Summary))
	}
	fmt.Fprintf(&sb, "  Run: %s  Feature: %s  Status: %s\n",
		report.Scope.ID, report.Scope.Feature, HumanStatus(report.Scope.Status))
	fmt.Fprintf(&sb, "  Tasks: %d\n", report.TaskCount)

	writeSectionHeader(&sb, "Gates")
	if report.PlanGate != nil {
		fmt.Fprintf(&sb, "  plan (run)  %s\n", humanGateStatus(report.PlanGate.Status))
		if report.PlanGate.Notes != "" {
			fmt.Fprintf(&sb, "    %s\n", report.PlanGate.Notes)
		}
	}
	atRisk := tasksAtRisk(report.Tasks)
	if len(atRisk) == 0 {
		sb.WriteString("  Tasks: toutes acceptables\n")
	} else {
		for _, t := range atRisk {
			if opts.Explain {
				fmt.Fprintf(&sb, "  %s  %s  %s  %s\n", t.TaskID, VerdictLabel(t.Verdict), scoreLabel(t.Score), HumanStatus(t.Status))
			} else {
				fmt.Fprintf(&sb, "  %s  %s  (%s)\n", t.TaskID, VerdictLabel(t.Verdict), HumanStatus(t.Status))
			}
		}
	}

	writeSectionHeader(&sb, "Risks")
	if len(atRisk) == 0 && (report.PlanGate == nil || !isBadGateStatus(report.PlanGate.Status)) {
		sb.WriteString("  (aucun)\n")
	} else {
		if report.PlanGate != nil && isBadGateStatus(report.PlanGate.Status) {
			fmt.Fprintf(&sb, "  • plan gate: %s\n", humanGateStatus(report.PlanGate.Status))
		}
		if len(atRisk) > 0 {
			fmt.Fprintf(&sb, "  • %d task(s) bloquée(s) ou à surveiller\n", len(atRisk))
		}
	}

	writeSectionHeader(&sb, "Next actions")
	writeRecommendations(&sb, report.NextActions)

	return sb.String()
}

func writeSectionHeader(sb *strings.Builder, title string) {
	if sb.Len() > 0 {
		sb.WriteByte('\n')
	}
	fmt.Fprintf(sb, "%s\n%s\n", title, sectionRule)
}

func writeRecommendations(sb *strings.Builder, recs []WorkTrustRecommendation) {
	var printed int
	for _, rec := range recs {
		if strings.TrimSpace(rec.Command) == "" && strings.TrimSpace(rec.Rationale) == "" {
			continue
		}
		if rec.Command != "" {
			fmt.Fprintf(sb, "  → %s\n", rec.Command)
		}
		if rec.Rationale != "" {
			fmt.Fprintf(sb, "    %s\n", rec.Rationale)
		}
		printed++
	}
	if printed == 0 {
		sb.WriteString("  (aucune)\n")
	}
}

func taskGateLines(report WorkTrustReport) []string {
	canonical := []string{"plan", gates.EnrichGateName, "governance", gates.HumanReviewGateName, gates.VerifyEvidenceGateName}
	statusByGate := map[string]string{}
	for _, d := range report.Dimensions {
		for _, g := range d.SourceGates {
			statusByGate[g] = gateStatusFromSummary(d.Summary, d.Status)
		}
	}
	for _, ev := range report.Evidences {
		if ev.Kind != "gate_history" {
			continue
		}
		name := strings.TrimPrefix(ev.Ref, "gate:")
		if name == ev.Ref {
			parts := strings.Split(ev.Summary, " ")
			if len(parts) > 0 {
				name = parts[0]
			}
		}
		if _, ok := statusByGate[name]; !ok && ev.Summary != "" {
			statusByGate[name] = humanGateStatus(ev.Summary)
		}
	}
	var lines []string
	for _, name := range canonical {
		status, ok := statusByGate[name]
		if !ok {
			status = "non évalué"
		}
		lines = append(lines, fmt.Sprintf("%-18s %s", name, status))
	}
	return lines
}

func gateStatusFromSummary(summary string, status DimensionStatus) string {
	s := strings.ToLower(summary)
	switch {
	case strings.Contains(s, "fail"):
		return "échec"
	case strings.Contains(s, "warn"):
		return "attention"
	case strings.Contains(s, "pass"):
		return "OK"
	case status == DimStatusUnevaluated:
		return "non évalué"
	case status == DimStatusFailed:
		return "échec"
	case status == DimStatusWeak:
		return "faible"
	default:
		return "OK"
	}
}

func tasksAtRisk(tasks []FeatureTaskSummary) []FeatureTaskSummary {
	var atRisk []FeatureTaskSummary
	for _, t := range tasks {
		if verdictRank(t.Verdict) >= verdictRank(VerdictRisky) {
			atRisk = append(atRisk, t)
		}
	}
	return atRisk
}

// VerdictLabel returns a human-readable trust verdict label.
func VerdictLabel(v Verdict) string {
	switch v {
	case VerdictTrusted:
		return "Fiable"
	case VerdictAcceptable:
		return "Acceptable"
	case VerdictRisky:
		return "À surveiller"
	case VerdictBlocked:
		return "Bloqué"
	default:
		return string(v)
	}
}

func verdictNarrative(v Verdict, _ string) string {
	switch v {
	case VerdictTrusted:
		return "Confiance élevée — le workflow peut avancer."
	case VerdictAcceptable:
		return "Confiance suffisante — quelques points à garder en tête."
	case VerdictRisky:
		return "Confiance limitée — vérifier les gates avant d'avancer."
	case VerdictBlocked:
		return "Action requise avant de continuer."
	default:
		return "État à inspecter."
	}
}

// HumanStatus returns a human-readable task status label.
func HumanStatus(status string) string {
	s := strings.TrimSpace(strings.ToLower(status))
	if s == "" {
		return "inconnu"
	}
	return strings.ReplaceAll(s, "_", " ")
}

func humanGateStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pass", "passed", "ok":
		return "OK"
	case "fail", "failed", "error":
		return "échec"
	case "warn", "warning":
		return "attention"
	default:
		if status == "" {
			return "non évalué"
		}
		return status
	}
}

func isBadGateStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "fail", "failed", "error", "warn", "warning":
		return true
	default:
		return false
	}
}

func humanSeverity(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical":
		return "critique"
	case "high":
		return "élevé"
	case "medium":
		return "modéré"
	case "low":
		return "faible"
	default:
		return s
	}
}

func dimensionStatusLabel(s DimensionStatus) string {
	switch s {
	case DimStatusStrong:
		return "fort"
	case DimStatusModerate:
		return "modéré"
	case DimStatusWeak:
		return "faible"
	case DimStatusFailed:
		return "échec"
	case DimStatusUnevaluated:
		return "n/a"
	default:
		return string(s)
	}
}

func topFindings(findings []WorkTrustFinding, n int) []WorkTrustFinding {
	if len(findings) == 0 {
		return nil
	}
	ranked := append([]WorkTrustFinding(nil), findings...)
	sort.SliceStable(ranked, func(i, j int) bool {
		return severityRank(ranked[i].Severity) > severityRank(ranked[j].Severity)
	})
	if len(ranked) > n {
		ranked = ranked[:n]
	}
	return ranked
}

func severityRank(s string) int {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	default:
		return 1
	}
}

// NextCommandForFeature returns the harmonized planner entrypoint for a feature.
func NextCommandForFeature(feature string) string {
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return "asa next"
	}
	return fmt.Sprintf("asa next --feature %s", feature)
}
