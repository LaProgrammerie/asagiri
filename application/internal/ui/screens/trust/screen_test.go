package trust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	got := Render(ViewModel{
		Trust: bus.TrustExplorerResult{
			Overall:      0.74,
			ResidualRisk: "medium",
			GateStatus:   "warn",
			GateReason:   "security confidence below profile target",
			Dimensions: []bus.TrustEvidenceDimension{
				{
					Label:         "Security",
					Score:         0.71,
					Findings:      []string{"no retry validation for invite_member"},
					Evidence:      []string{"security.flow check report"},
					CLIEquivalent: "asa verify trust onboarding --strict",
				},
			},
		},
		Model:   NewModel(),
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "trust.txt")
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

func TestRenderUnavailableTrust(t *testing.T) {
	got := Render(ViewModel{
		Trust: bus.TrustExplorerResult{
			Warning: "trust report unavailable",
		},
	})
	require.Contains(t, got, "Trust Summary")
	require.Contains(t, got, "Warning: trust report unavailable")
	require.Contains(t, got, "- unavailable")
}

