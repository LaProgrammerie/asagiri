package investigation_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/stretchr/testify/require"
)

func TestRunImpact(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rep, err := investigation.RunImpact(context.Background(), investigation.ImpactRequest{
		Flow: "onboarding", Change: "make invitations async", RepoRoot: dir, ProductID: "workspace-saas",
	})
	require.NoError(t, err)
	require.NotEmpty(t, rep.ID)
}
