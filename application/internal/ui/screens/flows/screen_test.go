package flows

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	got := Render(ViewModel{
		Flow: bus.FlowExplorerResult{
			FlowID: "onboarding",
			Steps: []bus.FlowStepDetail{
				{
					ID:         "invite_member",
					Label:      "invite_member",
					Status:     "running",
					API:        "POST /invitations",
					Service:    "InvitationService",
					Event:      "member.invited",
					Tests:      []string{"InvitationServiceTest"},
					Metrics:    []string{"invitation_success_rate"},
					TrustScore: 0.71,
					Risk:       "medium",
				},
			},
		},
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "flow.txt")
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

func TestRenderEmptyFlow(t *testing.T) {
	got := Render(ViewModel{
		Flow: bus.FlowExplorerResult{
			FlowID:  "onboarding",
			Warning: "flow graph unavailable",
		},
	})
	require.Contains(t, got, "Flow: onboarding")
	require.Contains(t, got, "- none")
}

