package contextopt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestCollectForPipelineUsesCandidatesOnly(t *testing.T) {
	dir := t.TempDir()
	noise := filepath.Join(dir, "noise.go")
	if err := os.WriteFile(noise, []byte("package noise\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "target.go")
	if err := os.WriteFile(target, []byte("package target\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.NewTestConfig("t")
	entries, err := CollectForPipeline(dir, "", cfg, CollectOpts{MaxFiles: 20}, []string{"target.go"})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.RelPath == "noise.go" {
			t.Fatal("full tree walk should not include noise.go when candidates set")
		}
	}
	found := false
	for _, e := range entries {
		if e.RelPath == "target.go" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected target.go in %+v", entries)
	}
}
