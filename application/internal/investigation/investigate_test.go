package investigation

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestRunOnTempRepo(t *testing.T) {
	dir := t.TempDir()
	if err := writeFile(filepath.Join(dir, "main.go"), "package main\nfunc HelloFeature() {}\n"); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(dir, "go.mod"), "module example.com/t\n\ngo 1.22\n"); err != nil {
		t.Fatal(err)
	}
	_ = exec.Command("git", "-C", dir, "init").Run()
	cfg := config.NewTestConfig("t")
	res, err := Run(t.Context(), dir, "HelloFeature", "", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.CandidateFiles) == 0 {
		t.Fatal("expected candidates")
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
