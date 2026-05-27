package api_test

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime/api"
	"github.com/stretchr/testify/require"
)

func TestServeUnix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sock := filepath.Join(dir, "runtime.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = api.ServeUnix(ctx, api.Options{RepoRoot: dir, SocketPath: sock})
	}()
	time.Sleep(200 * time.Millisecond)
	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Skip("unix socket not available")
	}
	_ = conn.Close()
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sock)
			},
		},
	}
	resp, err := client.Get("http://unix/v1/status")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = os.Remove(sock)
}
