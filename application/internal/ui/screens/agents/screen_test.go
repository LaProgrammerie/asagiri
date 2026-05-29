package agents

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	got := Render(ViewModel{
		Theatre: bus.AgentTheatreResult{
			Agents: []bus.AgentCard{
				{
					Role:            "implementer",
					AgentRef:        "cursor",
					Status:          "running",
					Task:            "Edit invitation service",
					FilesActive:     12,
					Hypothesis:      "missing retry in API client",
					TokensEstimated: 4200,
					CostEUR:         0.09,
					Duration:        3*time.Minute + 12*time.Second,
					LastOutput:      "added retry strategy and tests",
					Confidence:      0.78,
				},
				{
					Role:       "reviewer",
					AgentRef:   "codex",
					Status:     "waiting",
					Task:       "Review API contract changes",
					LastOutput: "waiting for implementation completion",
				},
			},
		},
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "agents.txt")
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
