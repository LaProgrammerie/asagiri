package agentadapter

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const defaultCodexCommand = "codex"

type codexAdapter struct{}

func (codexAdapter) Supports(providerType string, spec agentspec.Spec) bool {
	return providerType == config.ProviderTypeCodexCLI && targetCompatible(providerType, spec)
}

func (codexAdapter) Explain() string {
	return "Codex CLI : codex exec (ou args config), prompt orchestré sur stdin."
}

func (codexAdapter) Render(inv Invocation) (RenderedInvocation, error) {
	if !targetCompatible(inv.ProviderType, inv.Spec) {
		return RenderedInvocation{
			ProviderType: inv.ProviderType,
			AgentID:      inv.Spec.ID,
			SupportLevel: SupportUnsupported,
			Warnings:     targetWarnings(inv),
		}, fmt.Errorf("agentadapter: codex-cli non listé dans provider_targets")
	}
	prompt := strings.TrimSpace(inv.Prompt)
	if prompt == "" {
		return RenderedInvocation{}, fmt.Errorf("agentadapter: prompt vide")
	}
	cmd := commandOr(inv.Runtime.Command, defaultCodexCommand)
	args := append([]string(nil), inv.Runtime.Args...)
	if len(args) == 0 {
		args = []string{"exec"}
	}
	return baseRendered(inv, SupportInlinePrompt, cmd, args), nil
}
