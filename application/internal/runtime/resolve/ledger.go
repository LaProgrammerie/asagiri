package resolve

import (
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/configcache"
)

// EnrichLedgerParams applies registry fallback resolution metadata when enabled.
func EnrichLedgerParams(repoRoot string, cfg *config.Config, agentKey string, p *agentledger.Params) {
	if p == nil || cfg == nil || !cfg.Cloud.RegistryFallback {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	slug := strings.TrimSpace(cfg.Cloud.Project)
	if slug == "" {
		slug = "default"
	}
	res := ResolveAgent(Options{
		RepoRoot:    repoRoot,
		Config:      cfg,
		AgentKey:    agentKey,
		Cache:       configcache.NewStore(home),
		ProjectSlug: slug,
	})
	if res.Provider != "" {
		p.Provider = res.Provider
	}
	if res.Model != "" {
		p.Model = res.Model
	}
	if res.Role != "" && strings.TrimSpace(p.Role) == "" {
		p.Role = res.Role
	}
	p.ResolutionSource = res.Source
	p.RegistryAgentVersion = res.RegistryAgentVersion
	p.ConfigReleaseVersion = res.ConfigReleaseVersion
	p.ConfigContentHash = res.ConfigContentHash
}
