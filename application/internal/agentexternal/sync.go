package agentexternal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"gopkg.in/yaml.v3"
)

// Sync plans or applies external provider profile writes to explicitly configured paths.
func Sync(repoRoot string, cfg *config.Config, opts SyncOptions) (SyncReport, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return SyncReport{}, fmt.Errorf("agentexternal: repo_root requis")
	}

	loader := agentspec.NewLoader(repoRoot)
	specs, err := loader.LoadAll()
	if err != nil {
		return SyncReport{}, err
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })

	if id := strings.TrimSpace(opts.AgentID); id != "" {
		filtered := make([]agentspec.Spec, 0, 1)
		for _, spec := range specs {
			if spec.ID == id {
				filtered = append(filtered, spec)
				break
			}
		}
		if len(filtered) == 0 {
			return SyncReport{}, fmt.Errorf("agentexternal sync: agent %q introuvable", id)
		}
		specs = filtered
	}

	mode := "check"
	if opts.Write {
		mode = "write"
	}

	report := SyncReport{
		ReportVersion: SyncReportVersion,
		Mode:          mode,
		ReadOnly:      !opts.Write,
		Items:         make([]SyncItem, 0, len(specs)),
	}

	home, _ := os.UserHomeDir()
	for _, spec := range specs {
		configKey := strings.TrimSpace(opts.ConfigKey)
		if configKey == "" {
			configKey = spec.ID
		}
		item, wrote, err := syncOne(repoRoot, cfg, spec, configKey, home, opts)
		if err != nil {
			return report, err
		}
		if wrote {
			report.Wrote = true
		}
		report.Items = append(report.Items, item)
	}

	if hasSyncAction(report.Items, SyncActionConflict) {
		report.Hint = "Conflits détectés — relancer avec --force pour écraser le profil externe"
	} else if hasSyncPending(report.Items) && !opts.Write {
		report.Hint = "asa agents external sync --write"
	} else if report.Wrote {
		report.Hint = "Profils externes synchronisés"
	}
	return report, nil
}

func syncOne(repoRoot string, cfg *config.Config, spec agentspec.Spec, configKey, home string, opts SyncOptions) (SyncItem, bool, error) {
	target, err := resolveWritableTarget(spec, cfg, configKey, home)
	if err != nil {
		return SyncItem{}, false, err
	}

	item := SyncItem{
		AgentID:        spec.ID,
		ConfigKey:      configKey,
		Provider:       target.Provider,
		SupportLevel:   target.SupportLevel,
		LastSyncedHash: target.LastSyncedHash,
		Path:           target.DetectedPath,
	}

	if target.Status == StatusRejected || target.Status == StatusUnsupported || target.Status == StatusMissingPath {
		item.Action = SyncActionReject
		item.Message = target.Detail
		if item.Message == "" {
			item.Message = target.Status
		}
		return item, false, nil
	}

	configuredPath, pathSource, externalKind := resolveExternalPath(spec, cfg, configKey)
	if err := ensureAllowlistedAbsPath(configuredPath, target.DetectedPath, pathSource, home); err != nil {
		item.Action = SyncActionReject
		item.Message = err.Error()
		return item, false, nil
	}

	markdown := RenderProviderMarkdown(spec, externalKind)
	contentHash := contentHashBytes([]byte(markdown))
	item.ContentHash = contentHash

	installedHash, fileErr := fileSHA256(target.DetectedPath)
	if fileErr != nil {
		if !os.IsNotExist(fileErr) {
			item.Action = SyncActionReject
			item.Message = fileErr.Error()
			return item, false, nil
		}
		item.Action = SyncActionCreate
		item.Message = "fichier externe absent — création du profil provider"
		if !opts.Write {
			return item, false, nil
		}
		if err := writeExternalFile(target.DetectedPath, markdown); err != nil {
			return item, false, err
		}
		item.InstalledHash = contentHash
		specUpdated, specErr := patchSpecLastSyncedHash(repoRoot, spec.ID, contentHash, configuredPath, pathSource, externalKind)
		item.SpecUpdated = specUpdated
		if specErr != nil {
			item.Message += "; spec: " + specErr.Error()
		}
		return item, true, nil
	}

	item.InstalledHash = installedHash
	if installedHash == contentHash {
		item.Action = SyncActionSkip
		item.Message = "profil externe déjà à jour"
		return item, false, nil
	}

	item.Action = SyncActionConflict
	item.Message = "fichier externe modifié — hash différent du contenu Asagiri"
	if !opts.Force {
		if opts.Write {
			item.Message += " (non écrasé sans --force)"
		}
		return item, false, nil
	}

	item.Action = SyncActionUpdate
	item.Message = "écrasé depuis AgentSpec (--force)"
	if !opts.Write {
		return item, false, nil
	}
	if err := writeExternalFile(target.DetectedPath, markdown); err != nil {
		return item, false, err
	}
	item.InstalledHash = contentHash
	specUpdated, specErr := patchSpecLastSyncedHash(repoRoot, spec.ID, contentHash, configuredPath, pathSource, externalKind)
	item.SpecUpdated = specUpdated
	if specErr != nil {
		item.Message += "; spec: " + specErr.Error()
	}
	return item, true, nil
}

func writeExternalFile(absPath, content string) error {
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("agentexternal: mkdir %s: %w", dir, err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("agentexternal: écriture %s: %w", absPath, err)
	}
	return nil
}

func patchSpecLastSyncedHash(repoRoot, agentID, hash, configuredPath, pathSource, externalKind string) (bool, error) {
	specPath := filepath.Join(repoRoot, agentspec.RegistryDir, agentID+".yaml")
	data, err := os.ReadFile(specPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("registry %s absent — lancer asa agents sync --write", agentspec.RegistryDir+"/"+agentID+".yaml")
		}
		return false, err
	}

	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, fmt.Errorf("parse spec: %w", err)
	}

	ext, _ := doc["external"].(map[string]any)
	if ext == nil {
		ext = map[string]any{}
	}
	ext["last_synced_hash"] = hash
	if pathSource == "spec.external.path" {
		if p := strings.TrimSpace(configuredPath); p != "" {
			ext["path"] = p
		}
	} else if _, ok := ext["path"]; !ok {
		if p := strings.TrimSpace(configuredPath); p != "" {
			ext["path"] = p
		}
	}
	if k := strings.TrimSpace(externalKind); k != "" {
		ext["kind"] = k
	}
	doc["external"] = ext

	out, err := yaml.Marshal(doc)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(specPath, out, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func hasSyncAction(items []SyncItem, action string) bool {
	for _, it := range items {
		if it.Action == action {
			return true
		}
	}
	return false
}

func hasSyncPending(items []SyncItem) bool {
	for _, it := range items {
		switch it.Action {
		case SyncActionCreate, SyncActionUpdate, SyncActionConflict:
			return true
		}
	}
	return false
}

// HasBlockingSyncConflicts reports unresolved external sync conflicts.
func HasBlockingSyncConflicts(report SyncReport) bool {
	return hasSyncAction(report.Items, SyncActionConflict)
}
