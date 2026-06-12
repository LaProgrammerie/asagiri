package asagiri_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestRuntimeClientSession(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rt, err := asagiri.Connect(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = rt.Close() })

	sess, err := rt.StartSession("test-session", "p1", "onboarding")
	require.NoError(t, err)
	require.NotEmpty(t, sess.ID)
	require.NoError(t, rt.RunFlow(sess.ID, "onboarding"))
}
