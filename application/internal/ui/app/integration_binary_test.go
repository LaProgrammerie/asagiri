package app

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationBinaryDashboardNonTTY(t *testing.T) {
	bin := ensureASABinary(t)
	cmd := exec.Command(bin, "dashboard")
	cmd.Dir = repoRootForBinaryTest(t)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	require.NoError(t, cmd.Run())
	require.Contains(t, out.String(), "Usage:")
}

func TestIntegrationBinaryDashboardDryRunNonTTY(t *testing.T) {
	bin := ensureASABinary(t)
	cmd := exec.Command(bin, "dashboard", "--dry-run")
	cmd.Dir = repoRootForBinaryTest(t)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	require.NoError(t, cmd.Run())
	require.Contains(t, out.String(), "Usage:")
}

func TestIntegrationBinaryExists(t *testing.T) {
	bin := ensureASABinary(t)
	info, err := os.Stat(bin)
	require.NoError(t, err)
	require.False(t, info.IsDir())
}

func ensureASABinary(t *testing.T) string {
	t.Helper()
	root := repoRootForBinaryTest(t)
	bin := filepath.Join(root, "bin", "asa")
	if _, err := os.Stat(bin); err == nil {
		return bin
	}
	build := exec.Command("go", "build", "-o", bin, "./application/cmd/asa")
	build.Dir = root
	out, err := build.CombinedOutput()
	require.NoError(t, err, string(out))
	return bin
}

func repoRootForBinaryTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "application", "cmd", "asa")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
}
