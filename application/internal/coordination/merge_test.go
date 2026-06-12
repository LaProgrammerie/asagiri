package coordination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func TestMergeEvaluatorBlocksOnConflictsAndLowSecurity(t *testing.T) {
	eval := &coordination.MergeEvaluator{
		Require: []string{"trust_passed", "review_passed", "validation_passed"},
		BlockIf: []string{"unresolved_conflicts", "low_security_confidence"},
	}
	ok := eval.Evaluate(coordination.MergeContext{
		TrustPassed: true, ReviewPassed: true, ValidationPassed: true,
	})
	require.True(t, ok.Allowed)

	blocked := eval.Evaluate(coordination.MergeContext{
		TrustPassed: true, ReviewPassed: true, ValidationPassed: true,
		UnresolvedConflicts: 1,
	})
	require.False(t, blocked.Allowed)

	lowSec := eval.Evaluate(coordination.MergeContext{
		TrustPassed: true, ReviewPassed: true, ValidationPassed: true,
		SecurityConfidence: 0.5,
	})
	require.False(t, lowSec.Allowed)
}
