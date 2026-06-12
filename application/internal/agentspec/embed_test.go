package agentspec_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/stretchr/testify/require"
)

func TestListEmbeddedTemplates(t *testing.T) {
	templates, err := agentspec.ListEmbeddedTemplates()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(templates), 5)
	ids := make([]string, len(templates))
	for i, tpl := range templates {
		require.NotEmpty(t, tpl.ID)
		require.NotEmpty(t, tpl.Data)
		ids[i] = tpl.ID
	}
	require.Contains(t, ids, "dev")
	require.Contains(t, ids, "gate")
	require.Contains(t, ids, "enricher")
	require.Contains(t, ids, "governance")
	require.Contains(t, ids, "reviewer")
}
