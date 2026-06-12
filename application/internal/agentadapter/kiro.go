package agentadapter

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const defaultKiroCommand = "kiro"

type kiroAdapter struct{}

func (kiroAdapter) Supports(providerType string, spec agentspec.Spec) bool {
	return providerType == config.ProviderTypeKiroCLI && targetCompatible(providerType, spec)
}

func (kiroAdapter) Explain() string {
	return "Kiro CLI : kiro --cli (ou args config), prompt orchestré sur stdin — pas de sync ~/.kiro/agents."
}

func (kiroAdapter) Render(inv Invocation) (RenderedInvocation, error) {
	if !targetCompatible(inv.ProviderType, inv.Spec) {
		return RenderedInvocation{
			ProviderType: inv.ProviderType,
			AgentID:      inv.Spec.ID,
			SupportLevel: SupportUnsupported,
			Warnings:     targetWarnings(inv),
		}, fmt.Errorf("agentadapter: kiro-cli non listé dans provider_targets")
	}
	prompt := strings.TrimSpace(inv.Prompt)
	if prompt == "" {
		return RenderedInvocation{}, fmt.Errorf("agentadapter: prompt vide")
	}
	cmd := commandOr(inv.Runtime.Command, defaultKiroCommand)
	args := append([]string(nil), inv.Runtime.Args...)
	if len(args) == 0 {
		args = []string{"--cli"}
	}
	out := baseRendered(inv, SupportInlinePrompt, cmd, args)
	if strings.TrimSpace(inv.Runtime.Profile) != "" {
		out.Warnings = append(out.Warnings,
			fmt.Sprintf("config.profile %q ignoré — pas d'installation profil Kiro (T15)", inv.Runtime.Profile))
	}
	return out, nil
}
