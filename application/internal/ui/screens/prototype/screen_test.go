package prototype

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	got := Render(ViewModel{
		Pipeline: bus.PrototypePipelineResult{
			Product:        "workspace-saas",
			WireframeTitle: "workspace saas",
			WireframePath:  ".asagiri/products/workspace-saas/prototype/src/App.tsx",
			PipelineStage:  "flow",
			StagesDone:     []string{"wireframe", "journey", "flow"},
			Flow:           "workspace-onboarding",
			FlowExtraction: []bus.PrototypeFlowStep{
				{
					FlowID:   "workspace-onboarding",
					StepID:   "step-1",
					Action:   "click_get_started",
					Screen:   "landing",
					Next:     "signup",
					Contract: "POST /api/workspaces",
					Trust:    "pending",
					Metric:   "onboarding_completion_rate",
				},
				{
					FlowID:    "workspace-onboarding",
					StepID:    "step-2",
					Action:    "invite_member",
					Screen:    "signup",
					Next:      "dashboard",
					Contract:  "TODO:auth.signup",
					Trust:     "review-required",
					Metric:    "invitation_delivery_success_rate",
					Sensitive: true,
				},
			},
			SuggestedActions: []string{
				"asa contracts extract workspace-saas",
				"asa architecture derive workspace-saas",
			},
		},
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "prototype.txt")
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
