package product

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectArtifactsEmpty(t *testing.T) {
	root := t.TempDir()
	state := InspectArtifacts(root, "crm-artisans")
	if state.HasProductModel || state.HasPrototype || state.HasFlows || state.HasContracts || state.HasGeneratedSpecs || state.HasTasks {
		t.Fatalf("empty dir should have all false: %+v", state)
	}
	if len(state.MissingArtifacts) != 6 {
		t.Fatalf("expected 6 missing, got %d: %v", len(state.MissingArtifacts), state.MissingArtifacts)
	}
}

func TestInspectArtifactsPartial(t *testing.T) {
	root := t.TempDir()
	productDir := filepath.Join(root, ".asagiri", "products", "crm-artisans")
	if err := os.MkdirAll(productDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(productDir, "product.yaml"), []byte("name: crm-artisans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	protoDir := filepath.Join(productDir, "prototype")
	if err := os.MkdirAll(protoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(protoDir, "model.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	state := InspectArtifacts(root, "crm-artisans")
	if !state.HasProductModel || !state.HasPrototype {
		t.Fatalf("expected model+prototype: %+v", state)
	}
	if state.HasFlows || state.HasContracts {
		t.Fatalf("flows/contracts should be missing: %+v", state)
	}
}

func TestInspectArtifactsSpecsAndTasks(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, ".asagiri", "specs", "crm-artisans")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "tasks.yaml"), []byte("tasks: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte("# spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	state := InspectArtifacts(root, "crm-artisans")
	if !state.HasGeneratedSpecs || !state.HasTasks {
		t.Fatalf("expected specs and tasks: %+v", state)
	}
}
