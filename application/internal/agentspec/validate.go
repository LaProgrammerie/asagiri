package agentspec

import (
	"fmt"
	"path/filepath"
	"strings"
)

var knownRoles = map[string]struct{}{
	RoleDev:        {},
	RoleReviewer:   {},
	RoleEnricher:   {},
	RoleGovernance: {},
	RoleSpec:       {},
	RoleGate:       {},
	RoleUtility:    {},
}

var knownOutputFormats = map[string]struct{}{
	OutputAsagiriV1: {},
	OutputGateYAML:  {},
	OutputGateJSON:  {},
	OutputText:      {},
}

// Validate checks a parsed AgentSpec and optional source filename (for id/filename match).
func Validate(spec Spec, filename string) error {
	var errs []string

	id := strings.TrimSpace(spec.ID)
	if id == "" {
		errs = append(errs, "id: requis")
	} else if strings.Contains(id, " ") {
		errs = append(errs, "id: ne doit pas contenir d'espaces")
	}

	if strings.TrimSpace(spec.Version) == "" {
		errs = append(errs, "version: requise")
	}

	role := strings.TrimSpace(spec.Role)
	if role == "" {
		errs = append(errs, "role: requis")
	} else if _, ok := knownRoles[role]; !ok {
		errs = append(errs, fmt.Sprintf("role: %q inconnu (valeurs: dev, reviewer, enricher, governance, spec, gate, utility)", role))
	}

	if strings.TrimSpace(spec.SystemPrompt) == "" {
		errs = append(errs, "system_prompt: requis (non vide)")
	}

	format := strings.TrimSpace(spec.OutputContract.Format)
	if format == "" {
		errs = append(errs, "output_contract.format: requis")
	} else if _, ok := knownOutputFormats[format]; !ok {
		errs = append(errs, fmt.Sprintf("output_contract.format: %q inconnu (valeurs: asagiri-v1, gate-yaml, gate-json, text)", format))
	}

	for i, target := range spec.ProviderTargets {
		if strings.TrimSpace(target) == "" {
			errs = append(errs, fmt.Sprintf("provider_targets[%d]: ne doit pas être vide", i))
		}
	}

	if filename != "" && id != "" {
		base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		if base != id {
			errs = append(errs, fmt.Sprintf("id: %q ne correspond pas au fichier %q", id, base))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("agentspec invalide:\n  - %s", strings.Join(errs, "\n  - "))
}

// ValidateDuplicateIDs reports duplicate agent ids across a loaded set.
func ValidateDuplicateIDs(specs []Spec) error {
	seen := make(map[string]string, len(specs))
	var dups []string
	for _, spec := range specs {
		id := strings.TrimSpace(spec.ID)
		if prev, ok := seen[id]; ok {
			dups = append(dups, fmt.Sprintf("%q (%s et %s)", id, prev, spec.Source))
			continue
		}
		seen[id] = spec.Source
	}
	if len(dups) == 0 {
		return nil
	}
	return fmt.Errorf("agentspec: id dupliqué: %s", strings.Join(dups, "; "))
}
