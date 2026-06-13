package cloud_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/stretchr/testify/require"
)

func TestSaveLoadRemoveToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token")
	require.NoError(t, cloud.SaveToken(path, "secret-token"))
	token, err := cloud.LoadToken(path)
	require.NoError(t, err)
	require.Equal(t, "secret-token", token)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	require.NoError(t, cloud.RemoveToken(path))
	token, err = cloud.LoadToken(path)
	require.NoError(t, err)
	require.Empty(t, token)
}

func TestExpandPathHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	got, err := cloud.ExpandPath("~/tmp/asagiri-token")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, "tmp/asagiri-token"), got)
}

func TestRedactError(t *testing.T) {
	msg := cloud.RedactError(errString("Authorization Bearer sk-live-abc123"))
	require.NotContains(t, msg, "sk-live-abc123")
}

type errString string

func (e errString) Error() string { return string(e) }
