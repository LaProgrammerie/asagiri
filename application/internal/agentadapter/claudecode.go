package agentadapter

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const defaultClaudeCommand = "claude"

type claudeCodeAdapter struct{}

func (claudeCodeAdapter) Supports(providerType string, spec agentspec.Spec) bool {
	return providerType == config.ProviderTypeClaudeCode && targetCompatible(providerType, spec)
}

func (claudeCodeAdapter) Explain() string {
	return "Claude Code : --print + stream-json (args config), prompt orchestré sur stdin."
}

func (claudeCodeAdapter) Render(inv Invocation) (RenderedInvocation, error) {
	if !targetCompatible(inv.ProviderType, inv.Spec) {
		return RenderedInvocation{
			ProviderType: inv.ProviderType,
			AgentID:      inv.Spec.ID,
			SupportLevel: SupportUnsupported,
			Warnings:     targetWarnings(inv),
		}, fmt.Errorf("agentadapter: claude-code non listé dans provider_targets")
	}
	prompt := strings.TrimSpace(inv.Prompt)
	if prompt == "" {
		return RenderedInvocation{}, fmt.Errorf("agentadapter: prompt vide")
	}
	cmd := commandOr(inv.Runtime.Command, defaultClaudeCommand)
	args := append([]string(nil), inv.Runtime.Args...)
	if len(args) == 0 {
		args = []string{"--print", "--output-format", "stream-json"}
	}
	return baseRendered(inv, SupportNativeProfile, cmd, args), nil
}
