package coordination_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func TestConfigEscalationHandlerRetryAndEscalate(t *testing.T) {
	cfg := config.NewTestConfig("proj")
	cfg.Coordination.Retry.Implementation.MaxAttempts = 2
	cfg.Coordination.Escalation.AfterFailure = "investigator"
	cfg.Coordination.Escalation.AfterSecondFailure = "architecture_review"

	h := coordination.EscalationHandlerFromConfig(cfg)
	require.True(t, h.ShouldRetry(context.Background(), "n1", 1, 0))
	require.False(t, h.ShouldRetry(context.Background(), "n1", 2, 0))

	target, err := h.Escalate(context.Background(), "g1", "n1", 1)
	require.NoError(t, err)
	require.Equal(t, coordination.EscalationInvestigator, target)

	target, err = h.Escalate(context.Background(), "g1", "n1", 2)
	require.NoError(t, err)
	require.Equal(t, coordination.EscalationArchitectureReview, target)
}
