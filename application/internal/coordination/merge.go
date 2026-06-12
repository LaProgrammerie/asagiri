package coordination

import (
	"fmt"
	"strings"
)

// MergeContext carries gate inputs for merge evaluation (spec-my-D §16).
type MergeContext struct {
	TrustPassed          bool
	ReviewPassed         bool
	ValidationPassed     bool
	UnresolvedConflicts  int
	SecurityConfidence   float64
	LowSecurityThreshold float64
}

// MergeResult summarizes merge policy evaluation.
type MergeResult struct {
	Allowed bool     `json:"allowed"`
	Reasons []string `json:"reasons,omitempty"`
}

// MergeEvaluator applies coordination.merge require/block_if rules.
type MergeEvaluator struct {
	Require []string
	BlockIf []string
}

// Evaluate returns whether merge is allowed and blocking reasons.
func (m *MergeEvaluator) Evaluate(ctx MergeContext) MergeResult {
	if m == nil {
		return MergeResult{Allowed: true}
	}
	result := MergeResult{Allowed: true}
	lowThreshold := ctx.LowSecurityThreshold
	if lowThreshold <= 0 {
		lowThreshold = 0.7
	}

	for _, req := range m.Require {
		switch strings.TrimSpace(req) {
		case "trust_passed":
			if !ctx.TrustPassed {
				result.Allowed = false
				result.Reasons = append(result.Reasons, "trust_passed required")
			}
		case "review_passed":
			if !ctx.ReviewPassed {
				result.Allowed = false
				result.Reasons = append(result.Reasons, "review_passed required")
			}
		case "validation_passed":
			if !ctx.ValidationPassed {
				result.Allowed = false
				result.Reasons = append(result.Reasons, "validation_passed required")
			}
		}
	}

	for _, block := range m.BlockIf {
		switch strings.TrimSpace(block) {
		case "unresolved_conflicts":
			if ctx.UnresolvedConflicts > 0 {
				result.Allowed = false
				result.Reasons = append(result.Reasons, fmt.Sprintf(
					"unresolved_conflicts: %d", ctx.UnresolvedConflicts,
				))
			}
		case "low_security_confidence":
			if ctx.SecurityConfidence > 0 && ctx.SecurityConfidence < lowThreshold {
				result.Allowed = false
				result.Reasons = append(result.Reasons, fmt.Sprintf(
					"low_security_confidence: %.2f < %.2f", ctx.SecurityConfidence, lowThreshold,
				))
			}
		}
	}
	return result
}
