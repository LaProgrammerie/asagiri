package trust

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const productsRel = ".asagiri/products"

// ResolveProductID finds a product id from flow id by scanning .asagiri/products.
func ResolveProductID(repoRoot, flowID string) (string, error) {
	if flowID == "" {
		if id, err := defaultProductID(repoRoot); err == nil {
			return id, nil
		}
		return "", fmt.Errorf("trust: flow id required to resolve product")
	}
	dir := filepath.Join(repoRoot, productsRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fallbackProductID(repoRoot, flowID)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		productID := e.Name()
		flowsDir := filepath.Join(dir, productID, "flows")
		flowEntries, err := os.ReadDir(flowsDir)
		if err != nil {
			continue
		}
		for _, fe := range flowEntries {
			name := fe.Name()
			if !strings.HasSuffix(name, ".flow.yaml") {
				continue
			}
			stem := strings.TrimSuffix(name, ".flow.yaml")
			if stem == flowID || name == flowID {
				return productID, nil
			}
			raw, err := os.ReadFile(filepath.Join(flowsDir, name))
			if err != nil {
				continue
			}
			var meta struct {
				ID string `yaml:"id"`
			}
			if yaml.Unmarshal(raw, &meta) == nil && meta.ID == flowID {
				return productID, nil
			}
		}
	}
	if id, err := defaultProductID(repoRoot); err == nil {
		return id, nil
	}
	return "", fmt.Errorf("trust: no product found for flow %q", flowID)
}

func defaultProductID(repoRoot string) (string, error) {
	dir := filepath.Join(repoRoot, productsRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			return e.Name(), nil
		}
	}
	return "", fmt.Errorf("no products under %s", productsRel)
}

type productFlowMatch struct {
	FlowID    string
	ProductID string
}

// ResolveProductFlow locates a product flow id and product from a flow id or feature hint.
func ResolveProductFlow(repoRoot, hint string) (flowID, productID string, err error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return "", "", fmt.Errorf("product flow hint required")
	}
	matches := findProductFlowMatches(repoRoot, hint)
	switch len(matches) {
	case 0:
		return "", "", fmt.Errorf("no product flow matching %q", hint)
	case 1:
		return matches[0].FlowID, matches[0].ProductID, nil
	default:
		exact := filterExactFlowMatches(matches, hint)
		if len(exact) == 1 {
			return exact[0].FlowID, exact[0].ProductID, nil
		}
		return "", "", fmt.Errorf("ambiguous product flow for %q (%d matches)", hint, len(matches))
	}
}

func findProductFlowMatches(repoRoot, hint string) []productFlowMatch {
	dir := filepath.Join(repoRoot, productsRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	normHint := normalizeFlowHint(hint)
	var matches []productFlowMatch
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		productID := e.Name()
		flowsDir := filepath.Join(dir, productID, "flows")
		flowEntries, err := os.ReadDir(flowsDir)
		if err != nil {
			continue
		}
		for _, fe := range flowEntries {
			name := fe.Name()
			if !strings.HasSuffix(name, ".flow.yaml") {
				continue
			}
			stem := strings.TrimSuffix(name, ".flow.yaml")
			flowID := stem
			raw, readErr := os.ReadFile(filepath.Join(flowsDir, name))
			if readErr == nil {
				var meta struct {
					ID string `yaml:"id"`
				}
				if yaml.Unmarshal(raw, &meta) == nil && strings.TrimSpace(meta.ID) != "" {
					flowID = strings.TrimSpace(meta.ID)
				}
			}
			if flowMatchesHint(stem, flowID, normHint) {
				matches = append(matches, productFlowMatch{FlowID: flowID, ProductID: productID})
			}
		}
	}
	return matches
}

func filterExactFlowMatches(matches []productFlowMatch, hint string) []productFlowMatch {
	normHint := normalizeFlowHint(hint)
	var exact []productFlowMatch
	for _, m := range matches {
		if normalizeFlowHint(m.FlowID) == normHint {
			exact = append(exact, m)
		}
	}
	return exact
}

func normalizeFlowHint(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func flowMatchesHint(stem, flowID, normHint string) bool {
	if normHint == "" {
		return false
	}
	for _, candidate := range []string{stem, flowID} {
		if normalizeFlowHint(candidate) == normHint {
			return true
		}
	}
	return false
}

func fallbackProductID(repoRoot, flowID string) (string, error) {
	if _, err := os.Stat(filepath.Join(repoRoot, productsRel, "workspace-saas")); err == nil {
		return "workspace-saas", nil
	}
	return "", fmt.Errorf("trust: products directory missing under %s", repoRoot)
}
