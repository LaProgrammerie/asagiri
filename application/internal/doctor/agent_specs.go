package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentadapter"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentexternal"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

type agentSpecCollect struct {
	registry AgentRegistryInfo
	specs    []AgentSpecEntry
	drifts   []AgentDriftEntry
	last     *OrchestratedContextInfo
	checks   []Check
	actions  []Action
}

func collectAgentSpecs(repoRoot string, cfg *config.Config) agentSpecCollect {
	out := agentSpecCollect{}
	if cfg == nil {
		return out
	}

	loader := agentspec.NewLoader(repoRoot)
	registryPath := loader.AgentsDir()
	out.registry.Path = registryPath

	entries, listErr := listRegistryYAML(registryPath)
	out.registry.Present = listErr == nil && len(entries) > 0
	out.registry.FileCount = len(entries)
	out.registry.UsingEmbedded = !out.registry.Present

	if !out.registry.Present {
		out.registry.Status = StatusWarn
		out.registry.Detail = "registry absent — templates embarqués actifs ; matérialiser le disque avec asa agents sync --write"
		out.checks = append(out.checks, Check{
			ID:      "agent_registry",
			Status:  StatusWarn,
			Message: out.registry.Detail,
		})
		out.actions = appendUniqueAction(out.actions, Action{
			Title: "Matérialiser le registry AgentSpec",
			CLI:   "asa agents sync --write",
		})
	} else {
		out.registry.Status = StatusOK
		out.registry.Detail = fmt.Sprintf("%d fichier(s) AgentSpec", len(entries))
		out.checks = append(out.checks, Check{
			ID:      "agent_registry",
			Status:  StatusOK,
			Message: out.registry.Detail,
		})
	}

	diskByKey := loadDiskSpecs(loader, entries)
	keys := sortedAgentKeys(cfg)

	for _, key := range keys {
		entry, drifts := analyzeConfigAgent(repoRoot, cfg, key, diskByKey[key], out.registry.Present, entries)
		out.specs = append(out.specs, entry)
		out.drifts = append(out.drifts, drifts...)
	}

	out.drifts = append(out.drifts, missingWorkRoleSpecs(repoRoot, cfg, diskByKey, out.registry.Present)...)
	out.last = findLastOrchestratedContext(repoRoot)

	out.drifts = dedupeDrifts(out.drifts)
	for _, d := range out.drifts {
		status := StatusWarn
		if d.Kind == "invalid_spec" {
			status = StatusFail
		}
		out.checks = append(out.checks, Check{
			ID:      "agent_drift:" + d.ConfigKey + ":" + d.Kind,
			Status:  status,
			Message: d.Message,
		})
		if cli := strings.TrimSpace(d.FixCLI); cli != "" {
			out.actions = appendUniqueAction(out.actions, Action{
				Title: d.Message,
				CLI:   cli,
			})
		}
	}

	if len(out.drifts) > 0 {
		out.actions = appendUniqueAction(out.actions, Action{
			Title: "Aligner les AgentSpec du registry",
			CLI:   "asa agents sync --write",
		})
	}

	return out
}

func sortedAgentKeys(cfg *config.Config) []string {
	keys := make([]string, 0, len(cfg.Agents))
	for k := range cfg.Agents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func listRegistryYAML(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return files, nil
}

type diskSpecResult struct {
	spec agentspec.Spec
	err  error
}

func loadDiskSpecs(loader *agentspec.Loader, files []string) map[string]diskSpecResult {
	out := make(map[string]diskSpecResult, len(files))
	for _, name := range files {
		base := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		spec, err := loader.LoadDiskOnly(base)
		out[base] = diskSpecResult{spec: spec, err: err}
	}
	return out
}

func analyzeConfigAgent(repoRoot string, cfg *config.Config, key string, disk diskSpecResult, registryPresent bool, registryFiles []string) (AgentSpecEntry, []AgentDriftEntry) {
	entry := AgentSpecEntry{
		ConfigKey: key,
		Status:    StatusOK,
	}
	var drifts []AgentDriftEntry

	agentCfg, ok := cfg.Agents[key]
	if !ok {
		entry.Status = StatusFail
		entry.Detail = fmt.Sprintf("agents.%s absent", key)
		drifts = append(drifts, AgentDriftEntry{
			ConfigKey: key,
			Kind:      "missing_config_agent",
			Message:   entry.Detail,
			FixCLI:    "asa onboard --step agents",
		})
		return entry, drifts
	}

	providerType, merged, err := cfg.MergedAgentRuntime(key)
	if err != nil {
		entry.Status = StatusWarn
		entry.Detail = err.Error()
		drifts = append(drifts, AgentDriftEntry{
			ConfigKey: key,
			Kind:      "missing_provider",
			Message:   err.Error(),
			FixCLI:    "Éditer .asagiri/config.yaml (agents ou providers)",
		})
	} else {
		entry.ProviderType = strings.TrimSpace(providerType)
		if strings.TrimSpace(merged.Command) == "" && strings.TrimSpace(merged.Endpoint) == "" {
			entry.Status = StatusWarn
			entry.Detail = "commande provider absente"
			drifts = append(drifts, AgentDriftEntry{
				ConfigKey: key,
				Kind:      "missing_command",
				Message:   fmt.Sprintf("agents.%s : commande provider absente", key),
				FixCLI:    "Éditer .asagiri/config.yaml (providers." + strings.TrimSpace(agentCfg.Provider) + ")",
			})
		}
	}

	specFile := registryFileForKey(registryFiles, key)
	if disk.err != nil && specFile != "" {
		entry.Status = StatusFail
		entry.PromptSource = "invalid"
		entry.Detail = disk.err.Error()
		entry.Drift = append(entry.Drift, "invalid_spec")
		drifts = append(drifts, AgentDriftEntry{
			ConfigKey: key,
			Kind:      "invalid_spec",
			Message:   fmt.Sprintf("agents.%s : %s", key, disk.err.Error()),
			FixCLI:    fmt.Sprintf("Corriger %s", filepath.Join(repoRoot, agentspec.RegistryDir, specFile)),
		})
		return entry, drifts
	}

	diskSpec := disk.spec
	hasDisk := strings.TrimSpace(diskSpec.ID) != ""
	if hasDisk {
		entry.SpecID = diskSpec.ID
		entry.SpecVersion = diskSpec.Version
		entry.Role = diskSpec.Role
		entry.ContentHash = diskSpec.ContentHash
		entry.StoredHash = metadataContentHash(diskSpec)
		entry.PromptSource = "disk"
		entry.OutputFormat = diskSpec.OutputContract.Format
		entry.Detail = diskSpec.Source

		if stored := entry.StoredHash; stored != "" && stored != entry.ContentHash {
			msg := fmt.Sprintf("metadata.content_hash (%s) ≠ hash recalculé (%s)", truncateHash(stored), truncateHash(entry.ContentHash))
			entry.Drift = append(entry.Drift, "hash_mismatch")
			drifts = append(drifts, AgentDriftEntry{
				ConfigKey: key,
				Kind:      "hash_mismatch",
				Message:   fmt.Sprintf("agents.%s : %s", key, msg),
				FixCLI:    fmt.Sprintf("Mettre à jour metadata.content_hash dans %s ou asa agents sync --write --force --agent %s", agentspec.RegistryDir+"/"+key+".yaml", key),
			})
			if entry.Status == StatusOK {
				entry.Status = StatusWarn
			}
		}

		if expected := workRoleForAgentKey(cfg, key); expected != "" && diskSpec.Role != expected {
			msg := fmt.Sprintf("role spec %q ≠ rôle work attendu %q", diskSpec.Role, expected)
			entry.Drift = append(entry.Drift, "role_mismatch")
			drifts = append(drifts, AgentDriftEntry{
				ConfigKey: key,
				Kind:      "role_mismatch",
				Message:   fmt.Sprintf("agents.%s : %s", key, msg),
				FixCLI:    fmt.Sprintf("Corriger role dans %s", agentspec.RegistryDir+"/"+key+".yaml"),
			})
			entry.Status = StatusWarn
		}

		if entry.ProviderType != "" {
			level := providerSupportLevel(diskSpec, entry.ProviderType)
			entry.ProviderSupport = string(level)
			if level == agentadapter.SupportUnsupported {
				msg := fmt.Sprintf("provider.type %q non supporté par l'adapter pour ce AgentSpec", entry.ProviderType)
				entry.Drift = append(entry.Drift, "unsupported_provider")
				drifts = append(drifts, AgentDriftEntry{
					ConfigKey: key,
					Kind:      "unsupported_provider",
					Message:   fmt.Sprintf("agents.%s : %s", key, msg),
					FixCLI:    "Ajuster provider_targets ou config.providers",
				})
				entry.Status = StatusWarn
			}
		}

		if drift, ok := agentexternal.ExternalDrift(diskSpec, cfg, key); ok {
			entry.Drift = append(entry.Drift, drift.Kind)
			drifts = append(drifts, AgentDriftEntry{
				ConfigKey: key,
				Kind:      drift.Kind,
				Message:   fmt.Sprintf("agents.%s : %s", key, drift.Message),
				FixCLI:    drift.FixCLI,
			})
			if entry.Status == StatusOK {
				entry.Status = StatusWarn
			}
		}
	} else if registryPresent {
		entry.PromptSource = "missing"
		entry.Status = StatusWarn
		entry.Detail = fmt.Sprintf("spec disque absente (%s/%s.yaml)", agentspec.RegistryDir, key)
		drifts = append(drifts, AgentDriftEntry{
			ConfigKey: key,
			Kind:      "missing_spec",
			Message:   fmt.Sprintf("agents.%s configuré sans AgentSpec disque", key),
			FixCLI:    fmt.Sprintf("asa agents sync --write --agent %s", key),
		})
	} else {
		entry.PromptSource = "embedded"
		entry.Detail = "templates embarqués (registry absent)"
	}

	return entry, drifts
}

func registryFileExists(repoRoot, key string) bool {
	for _, ext := range []string{".yaml", ".yml"} {
		if _, err := os.Stat(filepath.Join(repoRoot, agentspec.RegistryDir, key+ext)); err == nil {
			return true
		}
	}
	return false
}

func registryFileForKey(files []string, key string) string {
	for _, name := range files {
		base := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		if base == key {
			return name
		}
	}
	return ""
}

func missingWorkRoleSpecs(repoRoot string, cfg *config.Config, diskByKey map[string]diskSpecResult, registryPresent bool) []AgentDriftEntry {
	if !registryPresent {
		return nil
	}
	var out []AgentDriftEntry
	for _, wr := range workRoles(cfg) {
		key := wr.agentKey
		if r, ok := diskByKey[key]; ok && r.err == nil && strings.TrimSpace(r.spec.ID) != "" {
			continue
		}
		if registryFileExists(repoRoot, key) {
			continue
		}
		out = append(out, AgentDriftEntry{
			ConfigKey: key,
			Kind:      "missing_work_role_spec",
			Message:   fmt.Sprintf("rôle %s : agents.%s sans AgentSpec disque", wr.role, key),
			FixCLI:    fmt.Sprintf("asa agents sync --write --agent %s", key),
		})
	}
	return out
}

type workRole struct {
	role     string
	agentKey string
}

func workRoles(cfg *config.Config) []workRole {
	return []workRole{
		{agentspec.RoleDev, cfg.WorkDevAgent()},
		{agentspec.RoleReviewer, cfg.WorkReviewerAgent()},
		{agentspec.RoleEnricher, cfg.WorkEnricherAgent()},
		{agentspec.RoleGovernance, cfg.GovernanceAgent()},
	}
}

func workRoleForAgentKey(cfg *config.Config, key string) string {
	for _, wr := range workRoles(cfg) {
		if wr.agentKey == key {
			return wr.role
		}
	}
	return ""
}

func providerSupportLevel(spec agentspec.Spec, providerType string) agentadapter.SupportLevel {
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

func metadataContentHash(spec agentspec.Spec) string {
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

func dedupeDrifts(in []AgentDriftEntry) []AgentDriftEntry {
	seen := make(map[string]struct{}, len(in))
	out := make([]AgentDriftEntry, 0, len(in))
	for _, d := range in {
		k := d.ConfigKey + "\x00" + d.Kind + "\x00" + d.Message
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, d)
	}
	return out
}

func enrichAgentsWithSpecs(agents []AgentInfo, specs []AgentSpecEntry) {
	byKey := make(map[string]AgentSpecEntry, len(specs))
	for _, s := range specs {
		byKey[s.ConfigKey] = s
	}
	for i := range agents {
		key := strings.TrimSpace(agents[i].LogicalID)
		if key == "" {
			continue
		}
		spec, ok := byKey[key]
		if !ok {
			continue
		}
		agents[i].SpecVersion = spec.SpecVersion
		agents[i].ContentHash = spec.ContentHash
		agents[i].StoredHash = spec.StoredHash
		agents[i].Drift = append([]string(nil), spec.Drift...)
		agents[i].PromptSource = spec.PromptSource
		agents[i].OutputFormat = spec.OutputFormat
		agents[i].ProviderSupport = spec.ProviderSupport
		if spec.Status == StatusWarn && agents[i].Status == StatusOK {
			agents[i].Status = StatusWarn
		}
		if spec.Status == StatusFail {
			agents[i].Status = StatusFail
		}
		if agents[i].Detail == "" && spec.Detail != "" {
			agents[i].Detail = spec.Detail
		}
	}
}

func findLastOrchestratedContext(repoRoot string) *OrchestratedContextInfo {
	logsRoot := filepath.Join(repoRoot, ".asagiri", "logs")
	entries, err := os.ReadDir(logsRoot)
	if err != nil {
		return nil
	}
	var bestPath string
	var bestTime time.Time
	for _, taskEnt := range entries {
		if !taskEnt.IsDir() {
			continue
		}
		agentsDir := filepath.Join(logsRoot, taskEnt.Name(), "agents")
		agentEntries, err := os.ReadDir(agentsDir)
		if err != nil {
			continue
		}
		for _, agentEnt := range agentEntries {
			if !agentEnt.IsDir() {
				continue
			}
			ctxPath := filepath.Join(agentsDir, agentEnt.Name(), "context.json")
			info, err := os.Stat(ctxPath)
			if err != nil {
				continue
			}
			if bestPath == "" || info.ModTime().After(bestTime) {
				bestPath = ctxPath
				bestTime = info.ModTime()
			}
		}
	}
	if bestPath == "" {
		return nil
	}
	data, err := os.ReadFile(bestPath)
	if err != nil {
		return nil
	}
	var ctx agentcontext.ExecutionContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil
	}
	rel, _ := filepath.Rel(repoRoot, bestPath)
	return &OrchestratedContextInfo{
		TaskID:    ctx.TaskID,
		AgentID:   ctx.AgentID,
		Feature:   ctx.Feature,
		Phase:     ctx.Phase,
		AgentHash: ctx.AgentHash,
		LogPath:   rel,
		UpdatedAt: bestTime.UTC().Format(time.RFC3339),
	}
}
