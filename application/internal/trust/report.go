package trust

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/safeid"
)

// ResidualRisk is a coarse risk label for the trust report (spec §6).
type ResidualRisk string

const (
	ResidualRiskUnknown ResidualRisk = "unknown"
	ResidualRiskLow     ResidualRisk = "low"
	ResidualRiskMedium  ResidualRisk = "medium"
	ResidualRiskHigh    ResidualRisk = "high"

	// ConfidenceUnavailableLabel is shown when no checks ran (spec §25, lot 1).
	ConfidenceUnavailableLabel = "n/a — no checks executed"
	// ConfidenceInferredCapLabel is shown for dimensions capped without a dedicated check (lot 2).
	ConfidenceInferredCapLabel = "≤50% (inferred cap)"
)

// TrustReport is the machine- and human-readable verification summary (spec §6).
type TrustReport struct {
	TrustID          string              `json:"trust_id"`
	GeneratedAt      string              `json:"generated_at"`
	Flow             string              `json:"flow,omitempty"`
	Branch           string              `json:"branch,omitempty"`
	Task             string              `json:"task,omitempty"`
	Repository       string              `json:"repository"`
	Checks           []VerificationCheck `json:"checks"`
	Confidence       confidence.Report   `json:"confidence"`
	BlastRadius      *BlastRadiusReport  `json:"blast_radius,omitempty"`
	Warnings         []string            `json:"warnings,omitempty"`
	ResidualRisk     ResidualRisk        `json:"residual_risk"`
	Gate             GateEvaluation      `json:"gate"`
	SuggestedReviews []SuggestedReview   `json:"suggested_reviews,omitempty"`
}

// WriteReport writes report.md and report.json under .asagiri/trust/<id>/ (spec §6).
func WriteReport(repoRoot, trustID string, report TrustReport) (mdPath, jsonPath string, err error) {
	if err := safeid.Validate(trustID); err != nil {
		return "", "", fmt.Errorf("write trust report: %w", err)
	}
	dir := filepath.Join(repoRoot, ".asagiri", "trust", trustID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create trust report dir: %w", err)
	}

	jsonPath = filepath.Join(dir, "report.json")
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal trust report json: %w", err)
	}
	if err := os.WriteFile(jsonPath, body, 0o644); err != nil {
		return "", "", fmt.Errorf("write trust report json: %w", err)
	}

	mdPath = filepath.Join(dir, "report.md")
	if err := os.WriteFile(mdPath, []byte(toTrustMarkdown(report)), 0o644); err != nil {
		return "", "", fmt.Errorf("write trust report markdown: %w", err)
	}
	return mdPath, jsonPath, nil
}

func toTrustMarkdown(r TrustReport) string {
	var sb strings.Builder
	sb.WriteString("# Trust Report\n\n")
	fmt.Fprintf(&sb, "- Trust ID: `%s`\n", r.TrustID)
	fmt.Fprintf(&sb, "- Generated: `%s`\n", r.GeneratedAt)
	if r.Flow != "" {
		fmt.Fprintf(&sb, "- Flow: `%s`\n", r.Flow)
	}
	if r.Branch != "" {
		fmt.Fprintf(&sb, "- Branch: `%s`\n", r.Branch)
	}
	if r.Task != "" {
		fmt.Fprintf(&sb, "- Task: `%s`\n", r.Task)
	}
	fmt.Fprintf(&sb, "- Residual risk: **%s**\n\n", r.ResidualRisk)

	sb.WriteString("## Confidence\n\n")
	sb.WriteString(formatConfidenceMarkdownLine("Architecture", confidence.DimensionArchitecture, r.Confidence.Architecture, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceMarkdownLine("Implementation", confidence.DimensionImplementation, r.Confidence.Implementation, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceMarkdownLine("Flow integrity", confidence.DimensionFlowIntegrity, r.Confidence.FlowIntegrity, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceMarkdownLine("Observability", confidence.DimensionObservability, r.Confidence.Observability, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceMarkdownLine("Security", confidence.DimensionSecurity, r.Confidence.Security, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceMarkdownLine("Regression", confidence.DimensionRegression, r.Confidence.Regression, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceMarkdownLine("Overall", confidence.Dimension("overall"), r.Confidence.Overall, r.Checks, r.Confidence))
	sb.WriteString("\n")

	if len(r.Confidence.UncoveredZones) > 0 {
		sb.WriteString("## Uncovered zones\n\n")
		for _, z := range r.Confidence.UncoveredZones {
			fmt.Fprintf(&sb, "- %s\n", z)
		}
		sb.WriteString("\n")
	}

	if len(r.Warnings) > 0 {
		sb.WriteString("## Warnings\n\n")
		for _, w := range r.Warnings {
			fmt.Fprintf(&sb, "- %s\n", w)
		}
		sb.WriteString("\n")
	}

	if r.BlastRadius != nil {
		sb.WriteString("## Blast Radius\n\n")
		sb.WriteString(formatBlastRadiusMarkdown(*r.BlastRadius))
		sb.WriteString("\n")
	}

	if len(r.Checks) > 0 {
		sb.WriteString("## Checks\n\n")
		for _, c := range r.Checks {
			fmt.Fprintf(&sb, "- `%s` [%s] %s\n", c.ID, c.Status, c.Name)
		}
		sb.WriteString("\n")
	}

	if len(r.SuggestedReviews) > 0 {
		sb.WriteString("## Suggested reviews\n\n")
		for _, rev := range r.SuggestedReviews {
			fmt.Fprintf(&sb, "- **%s**: %s\n", rev.Kind, rev.Reason)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Gate\n\n")
	fmt.Fprintf(&sb, "- Status: `%s`\n", r.Gate.Status)
	if r.Gate.Reason != "" {
		fmt.Fprintf(&sb, "- Reason: %s\n", r.Gate.Reason)
	}
	return sb.String()
}

// FormatTerminalSummary renders the lot 1 terminal UX skeleton (spec §25).
func FormatTerminalSummary(r TrustReport) string {
	var sb strings.Builder
	sb.WriteString("Asagiri Trust Engine\n")
	sb.WriteString("════════════════════\n")
	if r.Flow != "" {
		fmt.Fprintf(&sb, "Flow: %s\n", r.Flow)
	}
	if r.Branch != "" {
		fmt.Fprintf(&sb, "Branch: %s\n", r.Branch)
	}
	sb.WriteString("\nChecks\n")
	sb.WriteString("──────\n")
	if len(r.Checks) == 0 {
		sb.WriteString("(no checks executed)\n")
	} else {
		for _, c := range r.Checks {
			fmt.Fprintf(&sb, "%s %s\n", checkStatusGlyph(c.Status), c.Name)
		}
	}
	sb.WriteString("\nConfidence\n")
	sb.WriteString("──────────\n")
	sb.WriteString(formatConfidenceTerminalLine("Architecture", confidence.DimensionArchitecture, r.Confidence.Architecture, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceTerminalLine("Implementation", confidence.DimensionImplementation, r.Confidence.Implementation, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceTerminalLine("Flow integrity", confidence.DimensionFlowIntegrity, r.Confidence.FlowIntegrity, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceTerminalLine("Observability", confidence.DimensionObservability, r.Confidence.Observability, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceTerminalLine("Security", confidence.DimensionSecurity, r.Confidence.Security, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceTerminalLine("Regression", confidence.DimensionRegression, r.Confidence.Regression, r.Checks, r.Confidence))
	sb.WriteString(formatConfidenceTerminalLine("Overall", confidence.Dimension("overall"), r.Confidence.Overall, r.Checks, r.Confidence))

	if len(r.Confidence.UncoveredZones) > 0 {
		sb.WriteString("\nUncovered zones\n")
		sb.WriteString("───────────────\n")
		for _, z := range r.Confidence.UncoveredZones {
			fmt.Fprintf(&sb, "- %s\n", z)
		}
	}

	if len(r.Warnings) > 0 {
		sb.WriteString("\nWarnings\n")
		sb.WriteString("────────\n")
		for _, w := range r.Warnings {
			fmt.Fprintf(&sb, "- %s\n", w)
		}
	}
	if r.BlastRadius != nil {
		sb.WriteString("\nBlast Radius\n")
		sb.WriteString("────────────\n")
		sb.WriteString(formatBlastRadiusTerminal(*r.BlastRadius))
	}
	if len(r.SuggestedReviews) > 0 {
		sb.WriteString("\nSuggested reviews\n")
		sb.WriteString("─────────────────\n")
		for _, rev := range r.SuggestedReviews {
			fmt.Fprintf(&sb, "- %s: %s\n", rev.Kind, rev.Reason)
		}
	}
	fmt.Fprintf(&sb, "\nResidual risk: %s\n", r.ResidualRisk)
	fmt.Fprintf(&sb, "\nGate status: %s\n", strings.ToUpper(string(r.Gate.Status)))
	if r.Gate.Reason != "" {
		fmt.Fprintf(&sb, "Reason: %s\n", r.Gate.Reason)
	}
	return sb.String()
}

func noChecksExecuted(checks []VerificationCheck) bool {
	return len(checks) == 0
}

func formatConfidenceMarkdownLine(label string, dim confidence.Dimension, score float64, checks []VerificationCheck, conf confidence.Report) string {
	if noChecksExecuted(checks) {
		return fmt.Sprintf("- %s: %s\n", label, ConfidenceUnavailableLabel)
	}
	if dim != confidence.Dimension("overall") && confidenceDimensionInferred(conf, dim) {
		return fmt.Sprintf("- %s: %s\n", label, ConfidenceInferredCapLabel)
	}
	return fmt.Sprintf("- %s: %.0f%%\n", label, score*100)
}

func formatConfidenceTerminalLine(label string, dim confidence.Dimension, score float64, checks []VerificationCheck, conf confidence.Report) string {
	if noChecksExecuted(checks) {
		return fmt.Sprintf("%-16s %s\n", label+":", ConfidenceUnavailableLabel)
	}
	if dim != confidence.Dimension("overall") && confidenceDimensionInferred(conf, dim) {
		return fmt.Sprintf("%-16s %s\n", label+":", ConfidenceInferredCapLabel)
	}
	return fmt.Sprintf("%-16s %.2f\n", label+":", score)
}

func confidenceDimensionInferred(conf confidence.Report, dim confidence.Dimension) bool {
	for _, d := range conf.InferredDimensions {
		if d == string(dim) {
			return true
		}
	}
	return false
}

func checkStatusGlyph(status CheckStatus) string {
	switch status {
	case CheckStatusPassed:
		return "✓"
	case CheckStatusWarn:
		return "⚠"
	case CheckStatusFailed:
		return "✗"
	default:
		return "○"
	}
}

func formatBlastRadiusMarkdown(br BlastRadiusReport) string {
	return fmt.Sprintf(
		"Flows impacted: %d\nCritical APIs: %d\nShared modules: %d\nMigration risk: %s\nPublic contract risk: %s\n",
		br.FlowsImpacted, br.CriticalAPIs, br.SharedModules, br.MigrationRisk, br.PublicContractRisk,
	)
}

func formatBlastRadiusTerminal(br BlastRadiusReport) string {
	return fmt.Sprintf(
		"Flows impacted: %d\nCritical APIs: %d\nShared modules: %d\nMigration risk: %s\nPublic contract risk: %s\n",
		br.FlowsImpacted, br.CriticalAPIs, br.SharedModules, br.MigrationRisk, br.PublicContractRisk,
	)
}

// NewTrustReport builds a report with RFC3339 timestamp.
func NewTrustReport(scope VerificationScope, checks []VerificationCheck, conf confidence.Report, gate GateEvaluation, blast *BlastRadiusReport, suggested []SuggestedReview) TrustReport {
	warnings := append([]string{}, conf.Limits...)
	return TrustReport{
		TrustID:          scope.TrustID,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339Nano),
		Flow:             scope.Flow,
		Branch:           scope.Branch,
		Task:             scope.Task,
		Repository:       scope.RepoRoot,
		Checks:           checks,
		Confidence:       conf,
		BlastRadius:      blast,
		Warnings:         warnings,
		ResidualRisk:     ComputeResidualRisk(checks, conf),
		Gate:             gate,
		SuggestedReviews: suggested,
	}
}
