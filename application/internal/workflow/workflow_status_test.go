package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func TestStatusEmptyStore(t *testing.T) {
	repo := t.TempDir()
	cfg := config.NewTestConfig(filepath.Base(repo))
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o755))
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	svc := NewService(repo, cfg, store, true)
	runs, err := svc.Status(5)
	require.NoError(t, err)
	require.Empty(t, runs)
}

func TestSanitizeFeatureName(t *testing.T) {
	got := sanitize("My Feature!")
	if got == "" || got == "My Feature!" {
		t.Fatalf("got %q", got)
	}
}
