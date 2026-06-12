package replay

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	now := time.Date(2026, 5, 29, 8, 12, 0, 0, time.UTC)
	got := Render(ViewModel{
		Replay: bus.ReplayPackageResult{
			ReplayID:   "replay-001",
			CreatedAt:  now.Add(-10 * time.Minute),
			RepoBranch: "feature/spec-ui",
			RepoCommit: "abcdef1234567890",
			Mode:       "simulation",
			Artifacts:  []string{"graph/execution-graph.json", "runtime/events.jsonl"},
			Timeline: []bus.ReplayTimelineEvent{
				{Time: now, Type: "investigation.started"},
				{Time: now.Add(2 * time.Minute), Type: "graph.generated", Artifact: "graph/execution-graph.json"},
				{Time: now.Add(4 * time.Minute), Type: "implementation.started"},
			},
		},
		Model:   NewModel(),
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "replay.txt")
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
