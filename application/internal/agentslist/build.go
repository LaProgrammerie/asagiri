package agentslist

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentadapter"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Build collects a deterministic operator report for all available AgentSpec entries.
func Build(repoRoot string, cfg *config.Config) (Report, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return Report{}, fmt.Errorf("agentslist: repo_root requis")
	}

	loader := agentspec.NewLoader(repoRoot)
	specs, err := loader.LoadAll()
	if err != nil {
		return Report{}, err
	}

	registryPath := loader.AgentsDir()
	fileCount := 0
	if entries, listErr := os.ReadDir(registryPath); listErr == nil {
		for _, ent := range entries {
			if ent.IsDir() {
				continue
			}
			name := ent.Name()
			if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
				fileCount++
			}
		}
	}

	usingEmbedded := loader.UsingEmbeddedDefaults()
	report := Report{
		ReportVersion: ReportVersion,
		Registry: RegistryInfo{
			Path:                  agentspec.RegistryDir,
			Present:               fileCount > 0,
			FileCount:             fileCount,
			UsingEmbeddedDefaults: usingEmbedded,
		},
	}

	sort.Slice(specs, func(i, j int) bool {
		return specs[i].ID < specs[j].ID
	})

	configKeysBySpecID := indexConfigKeysBySpecID(cfg)
	for _, spec := range specs {
		report.Agents = append(report.Agents, buildEntry(spec, cfg, configKeysBySpecID[spec.ID]))
	}
	return report, nil
}

// Show returns one AgentSpec entry or an error if id is unknown.
func Show(repoRoot, id string, cfg *config.Config) (Entry, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Entry{}, fmt.Errorf("agentslist: id requis")
	}
	report, err := Build(repoRoot, cfg)
	if err != nil {
		return Entry{}, err
	}
	for _, entry := range report.Agents {
		if entry.ID == id {
			return entry, nil
		}
	}
	return Entry{}, fmt.Errorf("agentslist: agent %q introuvable", id)
}

func buildEntry(spec agentspec.Spec, cfg *config.Config, configKeys []string) Entry {
	entry := Entry{
		ID:              spec.ID,
		Role:            spec.Role,
		Version:         spec.Version,
		ContentHash:     spec.ContentHash,
		StoredHash:      storedContentHash(spec),
		Source:          sourceLabel(spec.Source),
		OutputFormat:    spec.OutputContract.Format,
		ProviderTargets: append([]string(nil), spec.ProviderTargets...),
	}

	if stored := entry.StoredHash; stored != "" && stored != entry.ContentHash {
		entry.Warnings = append(entry.Warnings,
			fmt.Sprintf("metadata.content_hash (%s) ≠ hash sémantique recalculé (%s)", truncateHash(stored), truncateHash(entry.ContentHash)))
	} else if stored == "" && entry.Source == "disk" {
		entry.Warnings = append(entry.Warnings, "metadata.content_hash absent — hash sémantique calculé au chargement")
	}

	entry.ProviderSupport = providerSupport(spec, cfg, configKeys)
	return entry
}

func indexConfigKeysBySpecID(cfg *config.Config) map[string][]string {
	out := map[string][]string{}
	if cfg == nil || len(cfg.Agents) == 0 {
		return out
	}
	keys := make([]string, 0, len(cfg.Agents))
	for k := range cfg.Agents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		specID := key
		out[specID] = append(out[specID], key)
	}
	return out
}

func providerSupport(spec agentspec.Spec, cfg *config.Config, configKeys []string) *ProviderSupport {
	targetLevels := targetSupportLevels(spec)
	if len(targetLevels) == 0 && cfg == nil {
		return nil
	}

	ps := &ProviderSupport{
		Targets: targetLevels,
		Summary: summarizeTargets(targetLevels),
	}

	if cfg == nil || len(configKeys) == 0 {
		return ps
	}

	configKey := configKeys[0]
	ps.ConfigKey = configKey
	providerType, _, err := cfg.MergedAgentRuntime(configKey)
	if err != nil {
		ps.Summary = ps.Summary + "; config: " + err.Error()
		return ps
	}
	ps.ProviderType = providerType
	ps.Level = string(supportForProvider(spec, providerType))
	ps.Summary = fmt.Sprintf("config %s → %s (%s); targets: %s",
		configKey, providerType, ps.Level, summarizeTargets(targetLevels))
	return ps
}

func targetSupportLevels(spec agentspec.Spec) map[string]string {
	matrix := agentadapter.SupportMatrix(spec)
	targets := spec.ProviderTargets
	if len(targets) == 0 {
		out := make(map[string]string, len(matrix))
		for _, pt := range sortedKeys(matrix) {
			out[pt] = string(matrix[pt])
		}
		return out
	}
	out := make(map[string]string, len(targets))
	for _, pt := range targets {
		pt = strings.TrimSpace(pt)
		if pt == "" {
			continue
		}
		if level, ok := matrix[pt]; ok {
			out[pt] = string(level)
		} else {
			out[pt] = string(agentadapter.SupportUnsupported)
		}
	}
	return out
}

func supportForProvider(spec agentspec.Spec, providerType string) agentadapter.SupportLevel {
	pt := strings.TrimSpace(providerType)
	if pt == "" {
		return agentadapter.SupportUnsupported
	}
	matrix := agentadapter.SupportMatrix(spec)
	if level, ok := matrix[pt]; ok {
		return level
	}
	return agentadapter.SupportUnsupported
}

func summarizeTargets(targets map[string]string) string {
	if len(targets) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(targets))
	for _, pt := range sortedKeys(targets) {
		parts = append(parts, fmt.Sprintf("%s:%s", pt, targets[pt]))
	}
	return strings.Join(parts, ", ")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sourceLabel(source string) string {
	if strings.TrimSpace(source) == "embedded" {
		return "embedded"
	}
	return "disk"
}

func storedContentHash(spec agentspec.Spec) string {
	if spec.Metadata == nil {
		return ""
	}
	v, ok := spec.Metadata["content_hash"]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func truncateHash(h string) string {
	h = strings.TrimSpace(h)
	if len(h) <= 12 {
		return h
	}
	return h[:12] + "…"
}
