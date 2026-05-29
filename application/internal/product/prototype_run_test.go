package product

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type stubDevRunner struct {
	dir string
}

func (s *stubDevRunner) StartDevServer(ctx context.Context, dir string) (*exec.Cmd, error) {
	s.dir = dir
	cmd := exec.CommandContext(ctx, "sleep", "30")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func TestRunPrototypeDryRun(t *testing.T) {
	repoRoot := t.TempDir()
	svc := NewService(repoRoot)
	productName, err := svc.CreatePrototype(CreatePrototypeOptions{
		Intent:  "demo",
		Product: "run-dry",
	})
	require.NoError(t, err)

	result, err := svc.RunPrototype(PrototypeRunOptions{Product: productName, DryRun: true})
	require.NoError(t, err)
	require.Equal(t, defaultPrototypeURL, result.URL)
	require.Equal(t, "npm run dev", result.Command)
	require.Contains(t, result.Dir, "prototype")
	require.Zero(t, result.PID)
}

func TestRunPrototypeWithTimeout(t *testing.T) {
	repoRoot := t.TempDir()
	svc := NewService(repoRoot)
	productName, err := svc.CreatePrototype(CreatePrototypeOptions{
		Intent:  "demo",
		Product: "run-timeout",
	})
	require.NoError(t, err)

	runner := &stubDevRunner{}
	result, err := svc.runPrototypeWithRunner(PrototypeRunOptions{
		Product: productName,
		Timeout: 200 * time.Millisecond,
	}, runner)
	require.NoError(t, err)
	require.NotZero(t, result.PID)
	require.Contains(t, runner.dir, "prototype")
}

func TestRunPrototypeMissingPrototype(t *testing.T) {
	svc := NewService(t.TempDir())
	_, err := svc.RunPrototype(PrototypeRunOptions{Product: "missing"})
	require.Error(t, err)
}
