package product

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareLayerDryRunNoWrites(t *testing.T) {
	root := t.TempDir()
	svc := NewService(root)
	before := dirSnapshot(root)

	res, err := svc.PrepareLayer(PrepareLayerOptions{
		Intent:  "Créer un CRM pour artisans",
		Product: "crm-artisans",
		DryRun:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Generated) == 0 {
		t.Fatal("dry-run should list planned outputs")
	}
	after := dirSnapshot(root)
	if before != after {
		t.Fatal("dry-run must not write files")
	}
}

func TestPrepareLayerCreatesMissingArtifacts(t *testing.T) {
	root := t.TempDir()
	svc := NewService(root)

	res, err := svc.PrepareLayer(PrepareLayerOptions{
		Intent:  "Créer un CRM pour artisans",
		Product: "crm-artisans",
		DryRun:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Generated) == 0 {
		t.Fatal("expected generated artifacts")
	}

	state := InspectArtifacts(root, "crm-artisans")
	if !state.HasProductModel || !state.HasPrototype || !state.HasFlows || !state.HasContracts || !state.HasGeneratedSpecs || !state.HasTasks {
		t.Fatalf("incomplete product layer: %+v", state)
	}
}

func TestPrepareLayerSkipsExisting(t *testing.T) {
	root := t.TempDir()
	svc := NewService(root)
	if _, err := svc.PrepareLayer(PrepareLayerOptions{
		Intent:  "CRM artisans",
		Product: "crm-artisans",
	}); err != nil {
		t.Fatal(err)
	}

	res, err := svc.PrepareLayer(PrepareLayerOptions{
		Intent:  "CRM artisans",
		Product: "crm-artisans",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Skipped) == 0 {
		t.Fatal("expected skipped steps on second run")
	}
}

func TestFormatLayerDryRun(t *testing.T) {
	out := FormatLayerDryRun(ProductArtifactState{
		ProductID:        "crm-artisans",
		MissingArtifacts: []string{"product model", "prototype"},
	}, false)
	if !strings.Contains(out, "Product-level intent detected") {
		t.Fatal("missing header")
	}
	if !strings.Contains(out, "Create product model") {
		t.Fatal("missing workflow steps")
	}
}

func TestFormatLayerDryRunPlanOnlyNote(t *testing.T) {
	out := FormatLayerDryRun(ProductArtifactState{ProductID: "crm-artisans"}, true)
	if !strings.Contains(out, "plan-only") {
		t.Fatalf("missing plan-only note: %s", out)
	}
}

func dirSnapshot(root string) string {
	var parts []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		parts = append(parts, path)
		return nil
	})
	return strings.Join(parts, "\n")
}
