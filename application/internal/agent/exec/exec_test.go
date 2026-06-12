package exec

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestExecutorDryRun(t *testing.T) {
	exec, err := New("cursor", config.Agent{
		Command: "cursor",
		Args:    []string{"agent", "run"},
	}, true)
	require.NoError(t, err)

	res, err := exec.Run(context.Background(), agent.RunRequest{
		Feature: "feature-a",
		TaskID:  "task-1",
		Prompt:  "test",
	})
	require.NoError(t, err)
	require.True(t, res.DryRun)
	require.Equal(t, 0, res.ExitCode)
	require.Contains(t, res.Command, "cursor")
}
