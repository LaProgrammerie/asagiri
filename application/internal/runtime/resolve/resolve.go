package resolve

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/configcache"
)

const (
	SourceLocal    = "local"
	SourceRegistry = "registry"
	SourceDefault  = "default"
)

// AgentResolution is the runtime outcome for provider/model selection.
type AgentResolution struct {
	AgentKey             string
	Role                 string
	Provider             string
	Model                string
	CapabilityKeys       []string
	RegistryAgentVersion int
	ConfigReleaseVersion int
	ConfigContentHash    string
	Source               string
	Warning              string
}

// Options controls registry fallback resolution.
type Options struct {
	RepoRoot    string
	Config      *config.Config
	AgentKey    string
	Cache       configcache.Store
	ProjectSlug string
	Now         time.Time
}

// ResolveAgent applies local → registry cache → defaults.
func ResolveAgent(opts Options) AgentResolution {
	agentKey := strings.TrimSpace(opts.AgentKey)
	if agentKey == "" {
		agentKey = "dev"
	}
	res := AgentResolution{AgentKey: agentKey, Source: SourceDefault}

	localProvider, localModel, localRole, hasLocalSpec := localAgentFields(opts.RepoRoot, agentKey)
	if hasLocalSpec {
		res.Provider = localProvider
		res.Model = localModel
		res.Role = localRole
		res.Source = SourceLocal
	}
	if opts.Config != nil {
		if p := agentProvider(opts.Config, agentKey); p != "" {
			res.Provider = p
			res.Source = SourceLocal
		}
		if m := strings.TrimSpace(opts.Config.AgentModel(agentKey)); m != "" {
			res.Model = m
			res.Source = SourceLocal
		}
	}

	if opts.Config == nil || !opts.Config.Cloud.RegistryFallback {
		return applyDefaults(opts.Config, res)
	}

	projectSlug := strings.TrimSpace(opts.ProjectSlug)
	if projectSlug == "" {
		projectSlug = strings.TrimSpace(opts.Config.Cloud.Project)
	}
	if projectSlug == "" {
		projectSlug = "default"
	}

	bundle, ptr, _, ok, err := opts.Cache.CurrentBundle(projectSlug)
	if err != nil {
		res.Warning = err.Error()
		return applyDefaults(opts.Config, res)
	}
	if !ok {
		res.Warning = "registry fallback activé mais aucun cache config local"
		return applyDefaults(opts.Config, res)
	}

	registryHit := resolveFromRegistry(bundle, agentKey, res)
	if registryHit.Source == SourceRegistry {
		if res.Source == SourceLocal && registryDrift(res, registryHit) {
			registryHit.Warning = "registry drift: configuration locale prioritaire"
			registryHit.Source = SourceLocal
		} else if res.Source != SourceLocal {
			res = registryHit
		} else {
			if res.Provider == "" {
				res.Provider = registryHit.Provider
			}
			if res.Model == "" {
				res.Model = registryHit.Model
			}
			if res.Role == "" {
				res.Role = registryHit.Role
			}
			if len(res.CapabilityKeys) == 0 {
				res.CapabilityKeys = registryHit.CapabilityKeys
			}
			res.RegistryAgentVersion = registryHit.RegistryAgentVersion
			res.ConfigReleaseVersion = registryHit.ConfigReleaseVersion
			res.ConfigContentHash = registryHit.ConfigContentHash
		}
	}

	if res.ConfigReleaseVersion == 0 {
		res.ConfigReleaseVersion = ptr.ReleaseVersion
	}
	if res.ConfigContentHash == "" {
		res.ConfigContentHash = ptr.ContentHash
	}

	if ageWarning := cacheAgeWarning(opts, ptr); ageWarning != "" {
		if res.Warning != "" {
			res.Warning += "; "
		}
		res.Warning += ageWarning
	}

	return applyDefaults(opts.Config, res)
}

func localAgentFields(repoRoot, agentKey string) (provider, model, role string, ok bool) {
	loader := agentspec.NewLoader(repoRoot)
	spec, err := loader.LoadDiskOnly(agentKey)
	if err != nil {
		return "", "", "", false
	}
	return "", "", strings.TrimSpace(spec.Role), true
}

func resolveFromRegistry(bundle cloud.ConfigBundle, agentKey string, base AgentResolution) AgentResolution {
	res := base
	registry, ok := bundle["registry"].(map[string]any)
	if !ok {
		return res
	}
	defs, ok := registry["agentDefinitions"].(map[string]any)
	if !ok {
		return res
	}
	rawDef, ok := defs[agentKey].(map[string]any)
	if !ok {
		return res
	}

	if res.Role == "" {
		res.Role = stringField(rawDef["role"])
	}
	if res.Provider == "" {
		res.Provider = stringField(rawDef["defaultProviderKey"])
	}
	if res.Model == "" {
		res.Model = stringField(rawDef["defaultModelKey"])
	}
	if active, ok := rawDef["activeVersion"].(map[string]any); ok {
		res.CapabilityKeys = stringSliceField(active["capabilityKeys"])
		if v, ok := active["version"].(float64); ok {
			res.RegistryAgentVersion = int(v)
		}
	}
	if prov, ok := bundle["provenance"].(map[string]any); ok {
		if v, ok := prov["releaseVersion"].(float64); ok {
			res.ConfigReleaseVersion = int(v)
		}
	}
	if hash, ok := bundle["contentHash"].(string); ok {
		res.ConfigContentHash = strings.TrimSpace(hash)
	}
	if res.Source != SourceLocal {
		res.Source = SourceRegistry
	}
	return res
}

func applyDefaults(cfg *config.Config, res AgentResolution) AgentResolution {
	if res.Provider == "" {
		res.Provider = agentProvider(cfg, res.AgentKey)
	}
	if res.Model == "" && cfg != nil {
		res.Model = strings.TrimSpace(cfg.AgentModel(res.AgentKey))
	}
	if res.Provider == "" && res.Model == "" && res.Source == SourceDefault {
		return res
	}
	if res.Source == "" {
		res.Source = SourceDefault
	}
	if res.Provider == "" && res.Model == "" {
		res.Source = SourceDefault
	}
	return res
}

func registryDrift(local, registry AgentResolution) bool {
	if local.Provider != "" && registry.Provider != "" && local.Provider != registry.Provider {
		return true
	}
	if local.Model != "" && registry.Model != "" && local.Model != registry.Model {
		return true
	}
	return false
}

func cacheAgeWarning(opts Options, ptr configcache.ActivePointer) string {
	if opts.Config == nil {
		return ""
	}
	maxAge, err := time.ParseDuration(strings.TrimSpace(opts.Config.Cloud.RegistryCacheMaxAge))
	if err != nil || maxAge <= 0 {
		return ""
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if now.Sub(ptr.FetchedAt) > maxAge {
		return fmt.Sprintf("cache config obsolète (%s)", now.Sub(ptr.FetchedAt).Truncate(time.Minute))
	}
	return ""
}

func stringField(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func stringSliceField(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

func agentProvider(cfg *config.Config, agentKey string) string {
	if cfg == nil || agentKey == "" {
		return ""
	}
	a, ok := cfg.Agents[agentKey]
	if !ok {
		return ""
	}
	return strings.TrimSpace(a.Provider)
}

// HomeDir resolves the user home for cache storage.
func HomeDir() (string, error) {
	return os.UserHomeDir()
}
