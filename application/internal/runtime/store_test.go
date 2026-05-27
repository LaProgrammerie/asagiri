package runtime

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimeStoreSessionBranchEvent(t *testing.T) {
	repo := t.TempDir()
	store, err := Open(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	st, err := StartDaemon(repo)
	require.NoError(t, err)
	require.True(t, st.Running)

	sess, err := store.CreateSession("onboarding-redesign", "workspace-saas", "onboarding")
	require.NoError(t, err)
	require.NotEmpty(t, sess.ID)

	branch, err := store.CreateBranch(sess.ID, "onboarding-enterprise", BranchFlow, "")
	require.NoError(t, err)
	require.Equal(t, sess.ID, branch.SessionID)

	_, err = store.EmitEvent("flow.started", "test", sess.ID, "onboarding", map[string]any{"step": "invite"})
	require.NoError(t, err)

	graph, err := store.BuildStateGraph()
	require.NoError(t, err)
	require.Len(t, graph.Sessions, 1)
	require.NotEmpty(t, graph.Events)

	require.NoError(t, StopDaemon(repo))
	st2, err := store.Status()
	require.NoError(t, err)
	require.False(t, st2.Running)

	require.FileExists(t, filepath.Join(repo, DefaultRelDir, "runtime.db"))
}
