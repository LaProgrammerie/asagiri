package asagiri_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime/api"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestHTTPClientSession(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	store, err := runtime.Open(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	srv := httptest.NewServer(api.NewServer(store).Handler())
	t.Cleanup(srv.Close)

	client := asagiri.ConnectHTTP(asagiri.HTTPOptions{BaseURL: srv.URL})
	sess, err := client.StartSession(context.Background(), "http-session", "p1", "onboarding")
	require.NoError(t, err)
	require.NotEmpty(t, sess.ID)
	require.NoError(t, client.RunFlow(context.Background(), sess.ID, "onboarding"))

	st, err := client.Status(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, st.Sessions, 1)
}
