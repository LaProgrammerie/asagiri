package explain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	got := Render(ViewModel{
		Explain: bus.ExplainResult{
			Subject:            "Why is this node blocked?",
			Question:           "Why is this node blocked?",
			SupportedQuestions: []string{"Why is this node blocked?"},
			Reasons:            []string{"Dependency investigate is not completed"},
			Evidence:      []string{"Node implement blocked by investigate"},
			Source:        "query-bus read-only",
			Alternatives:  []string{"asa graph", "asa trust"},
			CLIEquivalent: `asa explain --subject "Why is this node blocked?"`,
		},
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "explain.txt")
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

func TestRenderWarningState(t *testing.T) {
	got := Render(ViewModel{
		Explain: bus.ExplainResult{
			Subject: "blocked node",
			Warning: "no graph context available",
			Reasons: []string{"insufficient data"},
		},
		ShowCLI: true,
	})
	require.Contains(t, got, "Subject: blocked node")
	require.Contains(t, got, "Warning: no graph context available")
	require.Contains(t, got, "- insufficient data")
}

