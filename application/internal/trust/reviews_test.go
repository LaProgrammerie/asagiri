package trust

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

func TestSuggestReviewsSecurityAndProduct(t *testing.T) {
	report := TrustReport{
		Confidence:   confidence.Report{Overall: 0.5, Security: 0.6},
		ResidualRisk: ResidualRiskHigh,
		Checks: []VerificationCheck{
			{Type: CheckSecurity, Status: CheckStatusFailed, Findings: []Finding{
				{Severity: SeverityError, Category: "security.flow", Message: "auth missing"},
			}},
		},
	}
	reviews := SuggestReviews(report)
	require.NotEmpty(t, reviews)
	kinds := make(map[SuggestedReviewKind]bool)
	for _, r := range reviews {
		kinds[r.Kind] = true
	}
	require.True(t, kinds[ReviewSecurity])
	require.True(t, kinds[ReviewProduct])
}

func TestFormatSuggestedReviewsSection(t *testing.T) {
	report := TrustReport{
		SuggestedReviews: []SuggestedReview{{Kind: ReviewArchitecture, Reason: "blast radius"}},
		Gate:             GateEvaluation{Status: GateStatusPassed},
	}
	md := toTrustMarkdown(report)
	require.Contains(t, md, "## Suggested reviews")
	term := FormatTerminalSummary(report)
	require.Contains(t, term, "Suggested reviews")
}
