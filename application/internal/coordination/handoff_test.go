package coordination_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func TestDefaultHandoffBuilderPersistsYAML(t *testing.T) {
	repo := t.TempDir()
	builder := &coordination.DefaultHandoffBuilder{RepoRoot: repo}

	h, err := builder.Build(context.Background(), coordination.AgentResult{
		NodeID:     "implement-workspace",
		Role:       coordination.RoleInvestigator,
		TargetRole: coordination.RoleImplementer,
		Summary:    "onboarding invite failure likely caused by missing retry handling",
		Files:      []string{"src/Invitation/InvitationService.php"},
		Constraints: []string{
			"preserve public API",
		},
		Confidence: 0.78,
	})
	require.NoError(t, err)
	require.NotEmpty(t, h.ID)

	path := filepath.Join(repo, coordination.DefaultHandoffsRel, h.ID, "handoff.yaml")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var loaded coordination.Handoff
	require.NoError(t, yaml.Unmarshal(data, &loaded))
	require.Equal(t, coordination.RoleInvestigator, loaded.From)
	require.Equal(t, coordination.RoleImplementer, loaded.To)
	require.Equal(t, h.Summary, loaded.Summary)
	require.InDelta(t, 0.78, loaded.Confidence, 0.001)
}

func TestDefaultHandoffBuilderRequiresSummary(t *testing.T) {
	builder := &coordination.DefaultHandoffBuilder{RepoRoot: t.TempDir()}
	_, err := builder.Build(context.Background(), coordination.AgentResult{
		Role:       coordination.RoleInvestigator,
		TargetRole: coordination.RoleImplementer,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, coordination.ErrInvalidHandoff)
}
