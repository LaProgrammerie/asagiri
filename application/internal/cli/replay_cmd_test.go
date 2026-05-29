package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func replayConfigYAML() string {
	return `project:
  name: replay-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
replay:
  capture_prompts: true
  capture_runtime_events: true
  redact_secrets: true
`
}

func copyReplayCLIFixture(t *testing.T, repo string) {
	t.Helper()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	src := filepath.Join("..", "replay", "testdata", "replay", "basic-run")
	require.NoError(t, filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, "expected-artifacts.txt") {
			return nil
		}
		target := filepath.Join(repo, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, body, 0o644)
	}))
}

func TestReplayCreateCmd(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), replayConfigYAML())
	copyReplayCLIFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	defer func() { _ = os.Chdir(oldWd) }()

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{"replay", "create", "--from-graph", "graph-2026-05-29-test0001", "--include-runtime"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Asagiri Replay Engine")
	require.Contains(t, output.String(), "replay-")
}

func TestReplayRunOfflineCmd(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), replayConfigYAML())
	copyReplayCLIFixture(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	defer func() { _ = os.Chdir(oldWd) }()

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{"replay", "create", "--from-graph", "graph-2026-05-29-test0001"})
	require.NoError(t, root.Execute(), output.String())
	replayID := strings.TrimSpace(strings.Split(strings.Split(output.String(), "Replay:")[1], "\n")[0])

	output.Reset()
	root.SetArgs([]string{"replay", "run", replayID, "--offline"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Offline: true")
}

func TestReplayCompareCmd(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), replayConfigYAML())

	src := filepath.Join("..", "replay", "testdata", "replay", "divergence")
	require.NoError(t, filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(repo, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, body, 0o644)
	}))

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	defer func() { _ = os.Chdir(oldWd) }()

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{"replay", "create", "--from-graph", "graph-2026-05-29-test0001"})
	require.NoError(t, root.Execute(), output.String())
	idA := strings.TrimSpace(strings.Split(strings.Split(output.String(), "Replay:")[1], "\n")[0])

	output.Reset()
	root.SetArgs([]string{"replay", "create", "--from-graph", "graph-2026-05-29-test0002"})
	require.NoError(t, root.Execute(), output.String())
	idB := strings.TrimSpace(strings.Split(strings.Split(output.String(), "Replay:")[1], "\n")[0])

	output.Reset()
	root.SetArgs([]string{"replay", "compare", idA, idB})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Replay Comparison")
}

func TestCLIIntegrationReplay(t *testing.T) {
	TestReplayCreateCmd(t)
	TestReplayRunOfflineCmd(t)
	TestReplayCompareCmd(t)
}
