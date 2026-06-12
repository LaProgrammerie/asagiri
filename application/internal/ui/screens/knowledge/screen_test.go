package knowledge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	got := Render(ViewModel{
		Search: bus.KnowledgeSearchResult{
			Query: "invite_member",
			Matches: []bus.KnowledgeMatch{
				{
					ID:            "action:invite_member",
					Type:          "action",
					Name:          "invite_member",
					Path:          "application/internal/invitations/service.go",
					Score:         0.86,
					CLIEquivalent: `asa knowledge query "invite_member"`,
				},
			},
		},
		Model:   NewModel(),
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "knowledge.txt")
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

func TestRenderWarningAndNoMatch(t *testing.T) {
	got := Render(ViewModel{
		Search: bus.KnowledgeSearchResult{
			Query:   "missing-node",
			Warning: "knowledge graph unavailable",
		},
	})
	require.Contains(t, got, "Knowledge search: missing-node")
	require.Contains(t, got, "Warning: knowledge graph unavailable")
	require.Contains(t, got, "- no match")
}
