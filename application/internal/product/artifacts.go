package product

import (
	"os"
	"path/filepath"
	"strings"
)

// ProductArtifactState summarizes product-layer files on disk.
type ProductArtifactState struct {
	ProductID         string
	HasProductModel   bool
	HasPrototype      bool
	HasFlows          bool
	HasContracts      bool
	HasGeneratedSpecs bool
	HasTasks          bool
	MissingArtifacts  []string
}

// InspectArtifacts reads the filesystem for an existing product slug.
func InspectArtifacts(repoRoot, productID string) ProductArtifactState {
	productID = Slug(productID)
	repo := NewRepository(repoRoot)
	root := repo.productRoot(productID)

	state := ProductArtifactState{ProductID: productID}

	if fileExists(filepath.Join(root, "product.yaml")) {
		state.HasProductModel = true
	} else {
		state.MissingArtifacts = append(state.MissingArtifacts, "product model")
	}

	protoModel := filepath.Join(root, "prototype", "model.json")
	if fileExists(protoModel) || dirHasFiles(filepath.Join(root, "prototype")) {
		state.HasPrototype = true
	} else {
		state.MissingArtifacts = append(state.MissingArtifacts, "prototype")
	}

	if dirHasSuffixFiles(filepath.Join(root, "flows"), ".flow.yaml") {
		state.HasFlows = true
	} else {
		state.MissingArtifacts = append(state.MissingArtifacts, "flows")
	}

	if fileExists(filepath.Join(root, "contracts", "api.openapi.yaml")) {
		state.HasContracts = true
	} else {
		state.MissingArtifacts = append(state.MissingArtifacts, "contracts")
	}

	genReq := filepath.Join(root, "generated-specs", "requirements.md")
	specMD := filepath.Join(repo.specsRoot(productID), "spec.md")
	if fileExists(genReq) || fileExists(specMD) {
		state.HasGeneratedSpecs = true
	} else {
		state.MissingArtifacts = append(state.MissingArtifacts, "specs")
	}

	tasksYAML := filepath.Join(repo.specsRoot(productID), "tasks.yaml")
	if fileExists(tasksYAML) || dirHasTaskFiles(repo.tasksRoot(productID)) {
		state.HasTasks = true
	} else {
		state.MissingArtifacts = append(state.MissingArtifacts, "tasks")
	}

	return state
}

func (s ProductArtifactState) NeedsPreparation() bool {
	return len(s.MissingArtifacts) > 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirHasFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			return true
		}
	}
	return false
}

func dirHasSuffixFiles(dir, suffix string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), suffix) {
			return true
		}
	}
	return false
}

func dirHasTaskFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".json") {
			return true
		}
	}
	return false
}
