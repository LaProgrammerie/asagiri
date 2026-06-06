package factory

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/agent/claudecode"
	"github.com/LaProgrammerie/asagiri/application/internal/agent/exec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// NewFromConfig builds a runtime agent from config.agents + config.providers.
// Adapter selection uses provider.type only, never the agent or provider map key.
func NewFromConfig(name string, cfg *config.Config, dryRun bool) (agent.Agent, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config nil")
	}
	providerType, merged, err := cfg.MergedAgentRuntime(name)
	if err != nil {
		return nil, err
	}
	switch providerType {
	case config.ProviderTypeClaudeCode:
		return claudecode.New(name, merged, dryRun)
	case config.ProviderTypeExec,
		config.ProviderTypeKiroCLI,
		config.ProviderTypeCursorCLI,
		config.ProviderTypeCodexCLI,
		config.ProviderTypeGeminiCLI,
		config.ProviderTypeOllama:
		return exec.New(name, merged, dryRun)
	default:
		return nil, fmt.Errorf("agent %q: provider.type %q non supporté", name, providerType)
	}
}
