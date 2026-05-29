package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
	"gopkg.in/yaml.v3"
)

type productContext struct {
	productDir string
	flow       product.Flow
	bundle     analysis.Bundle
	bundleErr  error
}

func loadProductContext(scope Scope, deps Dependencies) (productContext, bool, error) {
	if scope.ProductID == "" {
		return productContext{}, true, nil
	}
	productDir := ProductDir(scope.RepoRoot, scope.ProductID)
	flowPath := resolveFlowPath(productDir, scope.Flow)
	raw, err := deps.ReadFile(flowPath)
	if err != nil {
		return productContext{}, false, fmt.Errorf("read flow %s: %w", scope.Flow, err)
	}
	flow, err := product.ParseFlowYAML(raw)
	if err != nil {
		return productContext{}, false, err
	}
	bundle, bundleErr := deps.LoadBundle(scope.RepoRoot, scope.ProductID)
	return productContext{
		productDir: productDir,
		flow:       flow,
		bundle:     bundle,
		bundleErr:  bundleErr,
	}, false, nil
}

func skippedLot3(scope Scope, start time.Time, typ, name, category, msg string) CheckResult {
	return CheckResult{
		ID:         checkID(typ, scope.TrustID),
		Name:       name,
		Type:       typ,
		Status:     statusSkipped,
		Confidence: 0,
		Findings: []Finding{{
			Severity: "info",
			Category: category,
			Message:  msg,
		}},
		Duration: time.Since(start),
	}
}

func contractExists(productDir, name string) bool {
	_, err := os.Stat(filepath.Join(productDir, "contracts", name))
	return err == nil
}

type permissionsContract struct {
	Roles map[string]struct {
		Permissions []string `yaml:"permissions"`
	} `yaml:"roles"`
}

func loadPermissionsContract(deps Dependencies, productDir string) (permissionsContract, error) {
	path := filepath.Join(productDir, "contracts", "permissions.yaml")
	raw, err := deps.ReadFile(path)
	if err != nil {
		return permissionsContract{}, err
	}
	var doc permissionsContract
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return permissionsContract{}, fmt.Errorf("parse permissions.yaml: %w", err)
	}
	return doc, nil
}

type observabilityContract struct {
	Requirements []string `yaml:"requirements"`
}

func loadObservabilityContract(deps Dependencies, productDir string) (observabilityContract, error) {
	path := filepath.Join(productDir, "contracts", "observability.yaml")
	raw, err := deps.ReadFile(path)
	if err != nil {
		return observabilityContract{}, err
	}
	var doc observabilityContract
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return observabilityContract{}, fmt.Errorf("parse observability.yaml: %w", err)
	}
	return doc, nil
}

type eventsContract struct {
	Events []string `yaml:"events"`
}

func loadEventsContract(deps Dependencies, productDir string) (eventsContract, error) {
	path := filepath.Join(productDir, "contracts", "events.yaml")
	raw, err := deps.ReadFile(path)
	if err != nil {
		return eventsContract{}, err
	}
	var doc eventsContract
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return eventsContract{}, fmt.Errorf("parse events.yaml: %w", err)
	}
	return doc, nil
}

func flowHasSensitiveStep(flow product.Flow) bool {
	for _, step := range flow.Steps {
		if step.Sensitive {
			return true
		}
	}
	return false
}

func unresolvedContractRefs(flow product.Flow) []string {
	var refs []string
	for _, step := range flow.Steps {
		ref := strings.TrimSpace(step.ContractRef)
		if ref == "" {
			continue
		}
		if strings.HasPrefix(ref, "TODO:") {
			refs = append(refs, ref)
		}
	}
	return refs
}
