package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultValidationCommands(t *testing.T) {
	dir := t.TempDir()
	cmds := validationLinesForRepo(dir)
	if len(cmds) != 2 {
		t.Fatalf("expected 2 cmds, got %v", cmds)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmds = validationLinesForRepo(dir)
	if len(cmds) < 2 {
		t.Fatalf("expected make lint, got %v", cmds)
	}
}

func TestParseValidationCommands(t *testing.T) {
	payload := `{"validation_commands":["go test ./...","go vet ./..."]}`
	got := parseValidationCommands(payload)
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}
