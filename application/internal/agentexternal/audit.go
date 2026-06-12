package agentexternal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentadapter"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const policyReadOnly = "external provider sync is opt-in; use asa agents external sync --write to push explicit paths only"

// Audit builds a read-only report of external provider targets and drift.
func Audit(repoRoot string, cfg *config.Config) (Report, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return Report{}, fmt.Errorf("agentexternal: repo_root requis")
	}

	loader := agentspec.NewLoader(repoRoot)
	specs, err := loader.LoadAll()
	if err != nil {
		return Report{}, err
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })

	home, _ := os.UserHomeDir()
	configKeysBySpec := indexConfigKeysBySpecID(cfg)

	report := Report{
		ReportVersion: ReportVersion,
		ReadOnly:      true,
		Policy:        policyReadOnly,
		Notes: []string{
			"Aucun scan de $HOME : seuls spec.external.path et agents.<id>.external_path sont résolus.",
			"asa agents sync (T20) matérialise le registry disque — pas les profils provider.",
			"Écriture provider : asa agents external sync --write (opt-in, chemins explicites uniquement).",
		},
	}

	for _, spec := range specs {
		keys := configKeysBySpec[spec.ID]
		if len(keys) == 0 {
			report.Targets = append(report.Targets, buildTarget(spec, "", cfg, home))
			continue
		}
		for _, key := range keys {
			report.Targets = append(report.Targets, buildTarget(spec, key, cfg, home))
		}
	}

	sort.Slice(report.Targets, func(i, j int) bool {
		a, b := report.Targets[i], report.Targets[j]
		if a.AgentID != b.AgentID {
			return a.AgentID < b.AgentID
		}
		return a.ConfigKey < b.ConfigKey
	})
	return report, nil
}

func buildTarget(spec agentspec.Spec, configKey string, cfg *config.Config, home string) ExternalTarget {
	target := ExternalTarget{
		AgentID:     spec.ID,
		ConfigKey:   strings.TrimSpace(configKey),
		DesiredHash: spec.ContentHash,
	}

	providerType, cliCmd, cliOK := resolveProvider(cfg, configKey)
	target.Provider = providerType
	target.CLICommand = cliCmd
	target.CLIAvailable = cliOK
	target.SupportLevel = string(supportForProvider(spec, providerType))

	configuredPath, pathSource, externalKind := resolveExternalPath(spec, cfg, configKey)
	target.ExternalKind = externalKind
	target.PathSource = pathSource
	target.ConfiguredPath = configuredPath

	if spec.External != nil {
		target.LastSyncedHash = strings.TrimSpace(spec.External.LastSyncedHash)
	}

	if configuredPath == "" {
		target.Status = StatusMissingPath
		target.Detail = "aucun chemin externe explicite (spec.external.path ou agents.external_path)"
		if !cliOK && providerType != "" && providerType != config.ProviderTypeExec {
			target.Status = StatusCLIMissing
			target.Detail = fmt.Sprintf("CLI %q absente du PATH", cliCmd)
		}
		return target
	}

	if strings.HasPrefix(strings.TrimSpace(configuredPath), "~/") && home == "" {
		target.Status = StatusPathUnexpanded
		target.Detail = "HOME indisponible — chemin ~ non résolu"
		return target
	}

	absPath := expandPath(configuredPath, home)
	target.DetectedPath = absPath
	target.Writable = pathWritable(absPath)

	if target.SupportLevel == string(agentadapter.SupportUnsupported) && providerType != "" {
		target.Status = StatusUnsupported
		target.Detail = fmt.Sprintf("provider.type %q non supporté pour cet AgentSpec", providerType)
	}

	if !cliOK && providerType != "" && providerType != config.ProviderTypeExec {
		target.Status = StatusCLIMissing
		target.Detail = fmt.Sprintf("CLI %q absente du PATH", cliCmd)
		return target
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			target.Status = StatusMissingFile
			target.Detail = "fichier externe absent"
			return target
		}
		target.Status = StatusMissingFile
		target.Detail = err.Error()
		return target
	}
	if info.IsDir() {
		target.Status = StatusMissingFile
		target.Detail = "chemin externe pointe vers un répertoire, fichier attendu"
		return target
	}

	hash, err := fileSHA256(absPath)
	if err != nil {
		target.Status = StatusMissingFile
		target.Detail = err.Error()
		return target
	}
	target.InstalledHash = hash

	target.Status = StatusOK
	target.Detail = "fichier externe présent"
	if target.LastSyncedHash != "" && hash != target.LastSyncedHash {
		target.Status = StatusDrift
		target.Detail = "installed_hash ≠ spec.external.last_synced_hash"
	} else if target.LastSyncedHash == "" && hash != spec.ContentHash {
		target.Status = StatusDrift
		target.Detail = "installed_hash ≠ hash sémantique AgentSpec (last_synced_hash absent)"
	}
	return target
}

func resolveExternalPath(spec agentspec.Spec, cfg *config.Config, configKey string) (path, source, kind string) {
	if spec.External != nil {
		if p := strings.TrimSpace(spec.External.Path); p != "" {
			return p, "spec.external.path", strings.TrimSpace(spec.External.Kind)
		}
		kind = strings.TrimSpace(spec.External.Kind)
	}
	if cfg != nil && configKey != "" {
		if a, err := cfg.LookupAgent(configKey); err == nil {
			if p := strings.TrimSpace(a.ExternalPath); p != "" {
				return p, "config.agents.external_path", kind
			}
		}
	}
	return "", "", kind
}

func resolveProvider(cfg *config.Config, configKey string) (providerType, cliCmd string, available bool) {
	if cfg == nil || strings.TrimSpace(configKey) == "" {
		return "", "", false
	}
	providerType, merged, err := cfg.MergedAgentRuntime(configKey)
	if err != nil {
		return "", "", false
	}
	cliCmd = strings.TrimSpace(merged.Command)
	if cliCmd == "" {
		return providerType, "", false
	}
	_, err = exec.LookPath(cliCmd)
	return providerType, cliCmd, err == nil
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

func expandPath(raw, home string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "~/") {
		if home == "" {
			return raw
		}
		return filepath.Clean(filepath.Join(home, strings.TrimPrefix(raw, "~/")))
	}
	if strings.HasPrefix(raw, "~") && len(raw) > 1 && raw[1] == filepath.Separator {
		if home == "" {
			return raw
		}
		return filepath.Clean(filepath.Join(home, raw[1:]))
	}
	return filepath.Clean(raw)
}

func pathWritable(path string) bool {
	path = filepath.Clean(path)
	if path == "" {
		return false
	}
	if st, err := os.Stat(path); err == nil {
		return st.Mode().Perm()&0200 != 0
	} else if os.IsNotExist(err) {
		dir := filepath.Dir(path)
		for {
			if st, err2 := os.Stat(dir); err2 == nil {
				return st.IsDir() && st.Mode().Perm()&0200 != 0
			} else if !os.IsNotExist(err2) {
				return false
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				return false
			}
			dir = parent
		}
	}
	return false
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hashBytes(data), nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
