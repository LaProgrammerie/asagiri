package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeExampleConfig(t *testing.T, repo string) {
	t.Helper()
	af := filepath.Join(repo, ".agentflow")
	if err := os.MkdirAll(af, 0o755); err != nil {
		t.Fatal(err)
	}
	example := filepath.Join(af, "config.yaml.example")
	src := filepath.Join("..", "..", "..", ".agentflow", "config.yaml.example")
	data, err := os.ReadFile(src)
	if err != nil {
		// fallback minimal example
		data = []byte(`project:
  name: test
state:
  backend: sqlite
  path: .agentflow/state.sqlite
worktrees:
  base_path: .agentflow/worktrees
`)
	}
	if err := os.WriteFile(example, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInitWithoutGit(t *testing.T) {
	dir := t.TempDir()
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init"})
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	if err := root.Execute(); err == nil {
		t.Fatal("expected error without git repo")
	}
}

func TestInitAndDoubleInit(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)

	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatalf("first init: %v", err)
	}

	cfgPath := filepath.Join(dir, ".agentflow", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	first := string(data)

	if err := root.Execute(); err != nil {
		t.Fatalf("second init: %v", err)
	}
	data2, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data2) != first {
		t.Fatal("config.yaml was modified on second init")
	}

	for _, sub := range []string{"runs", "tasks", "logs", "worktrees"} {
		if _, err := os.Stat(filepath.Join(dir, ".agentflow", sub)); err != nil {
			t.Fatalf("missing %s: %v", sub, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".agentflow", "state.sqlite")); err != nil {
		t.Fatal("state.sqlite missing:", err)
	}
}

func TestDoctorBeforeAndAfterInit(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)

	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"doctor"})
	if err := root.Execute(); err == nil {
		t.Fatal("doctor should fail before init")
	}

	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	root.SetArgs([]string{"doctor"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doctor after init: %v", err)
	}
}
