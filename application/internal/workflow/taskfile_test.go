package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/plan"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

func TestCanonicalTaskYAMLRoundTrip(t *testing.T) {
	canonical := planToCanonical("feat", plan.Task{ID: "feat-001", Title: "A"})
	body, err := yaml.Marshal(canonical)
	if err != nil {
		t.Fatal(err)
	}
	var back asagiri.Task
	if err := yaml.Unmarshal(body, &back); err != nil {
		t.Fatal(err)
	}
	if back.ID != "feat-001" {
		t.Fatalf("id %q", back.ID)
	}
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	svc := &Service{repoRoot: repo}
	if err := svc.persistCanonicalTaskFiles("feat", []asagiri.Task{canonical}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".asagiri", "tasks", "feat", "feat-001.yaml")); err != nil {
		t.Fatal(err)
	}
}
