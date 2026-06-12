package trust

import (
	"fmt"
	"strings"
)

// SuggestedReviewKind identifies an automated review lane (spec-my-B §20).
type SuggestedReviewKind string

const (
	ReviewArchitecture  SuggestedReviewKind = "architecture"
	ReviewSecurity      SuggestedReviewKind = "security"
	ReviewObservability SuggestedReviewKind = "observability"
	ReviewPerformance   SuggestedReviewKind = "performance"
	ReviewProduct       SuggestedReviewKind = "product"
)

// SuggestedReview is a recommended follow-up review after verification.
type SuggestedReview struct {
	Kind   SuggestedReviewKind `json:"kind"`
	Reason string              `json:"reason"`
}

const lowConfidenceOverall = 0.7

// SuggestReviews derives review orchestration hints from checks, confidence, and blast radius (§20).
func SuggestReviews(report TrustReport) []SuggestedReview {
	var out []SuggestedReview
	add := func(kind SuggestedReviewKind, reason string) {
		for _, existing := range out {
			if existing.Kind == kind {
				return
			}
		}
		out = append(out, SuggestedReview{Kind: kind, Reason: reason})
	}

	if report.Confidence.Overall > 0 && report.Confidence.Overall < lowConfidenceOverall {
		add(ReviewProduct, fmt.Sprintf("overall confidence %.0f%% below %.0f%%", report.Confidence.Overall*100, lowConfidenceOverall*100))
	}
	if report.ResidualRisk == ResidualRiskHigh || report.ResidualRisk == ResidualRiskMedium {
		add(ReviewProduct, fmt.Sprintf("residual risk is %s", report.ResidualRisk))
	}

	if br := report.BlastRadius; br != nil {
		if br.FlowsImpacted >= 2 || br.CriticalAPIs >= 1 || br.PublicContractRisk == "high" {
			add(ReviewArchitecture, "elevated blast radius on flows or public contracts")
		}
	}

	for _, c := range report.Checks {
		switch c.Type {
		case CheckArchitecture:
			if c.Status == CheckStatusFailed || c.Status == CheckStatusWarn {
				add(ReviewArchitecture, checkReason(c, "architecture check"))
			}
		case CheckSecurity:
			if c.Status == CheckStatusFailed || hasSeverity(c.Findings, SeverityError, SeverityCritical) {
				add(ReviewSecurity, checkReason(c, "security check"))
			}
		case CheckObservability:
			if c.Status == CheckStatusFailed || c.Status == CheckStatusWarn {
				add(ReviewObservability, checkReason(c, "observability check"))
			}
		case CheckPerformance:
			if c.Status == CheckStatusFailed || c.Status == CheckStatusWarn {
				add(ReviewPerformance, checkReason(c, "performance check"))
			}
		case CheckFlows:
			if c.Status == CheckStatusFailed || hasCategory(c.Findings, "flow.integrity", "flow.security") {
				add(ReviewArchitecture, checkReason(c, "flow integrity"))
				if hasCategory(c.Findings, "flow.security") {
					add(ReviewSecurity, "flow security findings")
				}
			}
		case CheckContracts, CheckBackwardCompatibility:
			if c.Status == CheckStatusFailed || c.Status == CheckStatusWarn {
				add(ReviewArchitecture, checkReason(c, string(c.Type)+" check"))
			}
		}
	}

	if report.Confidence.Security > 0 && report.Confidence.Security < lowConfidenceOverall {
		add(ReviewSecurity, fmt.Sprintf("security confidence %.0f%%", report.Confidence.Security*100))
	}
	if report.Confidence.Observability > 0 && report.Confidence.Observability < lowConfidenceOverall {
		add(ReviewObservability, fmt.Sprintf("observability confidence %.0f%%", report.Confidence.Observability*100))
	}
	if report.Confidence.Architecture > 0 && report.Confidence.Architecture < lowConfidenceOverall {
		add(ReviewArchitecture, fmt.Sprintf("architecture confidence %.0f%%", report.Confidence.Architecture*100))
	}
	return out
}

func checkReason(c VerificationCheck, label string) string {
	if len(c.Findings) > 0 {
		return fmt.Sprintf("%s: %s", label, c.Findings[0].Message)
	}
	return fmt.Sprintf("%s status %s", label, c.Status)
}

func hasSeverity(findings []Finding, severities ...Severity) bool {
	want := make(map[Severity]bool, len(severities))
	for _, s := range severities {
		want[s] = true
	}
	for _, f := range findings {
		if want[f.Severity] {
			return true
		}
	}
	return false
}

func hasCategory(findings []Finding, categories ...string) bool {
	for _, f := range findings {
		for _, c := range categories {
			if strings.HasPrefix(f.Category, c) {
				return true
			}
		}
	}
	return false
}
