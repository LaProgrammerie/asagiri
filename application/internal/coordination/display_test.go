package coordination_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/stretchr/testify/require"
)

func TestFormatMultiAgentRuntimeGolden(t *testing.T) {
	view := coordination.MultiAgentRuntimeView{
		Pipeline: []coordination.RuntimeDisplayStep{
			{Role: coordination.RoleInvestigator, Status: coordination.RuntimeDisplayCompleted},
			{Role: coordination.RoleArchitect, Status: coordination.RuntimeDisplayCompleted},
			{Role: coordination.RoleImplementer, Status: coordination.RuntimeDisplayRunning},
			{Role: coordination.RoleReviewer, Status: coordination.RuntimeDisplayPending},
			{Role: coordination.RoleValidator, Status: coordination.RuntimeDisplayPending},
		},
		Agents: []coordination.ActiveAgent{
			{Role: coordination.RoleImplementer, AgentRef: "cursor"},
			{Role: coordination.RoleReviewer, AgentRef: "claude"},
			{Role: coordination.RoleValidator, AgentRef: "local"},
		},
		CurrentEUR: 0.11,
		BudgetEUR:  1.00,
		Warnings: []string{
			"reviewer pending",
			"trust validation required before merge",
		},
	}

	got := coordination.FormatMultiAgentRuntime(view)
	golden := filepath.Join("testdata", "multi_agent_runtime.txt")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
	}

	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) {
		t.Fatalf("golden %s missing; run with UPDATE_GOLDEN=1", golden)
	}
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}
