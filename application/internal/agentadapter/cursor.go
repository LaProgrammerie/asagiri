package agentadapter

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const defaultCursorCommand = "cursor-agent"

type cursorAdapter struct{}

func (cursorAdapter) Supports(providerType string, spec agentspec.Spec) bool {
	return providerType == config.ProviderTypeCursorCLI && targetCompatible(providerType, spec)
}

func (cursorAdapter) Explain() string {
	return "Cursor CLI : commande runtime (défaut cursor-agent), args config, prompt orchestré sur stdin — pas d'installation de profil Cursor."
}

func (cursorAdapter) Render(inv Invocation) (RenderedInvocation, error) {
	if !targetCompatible(inv.ProviderType, inv.Spec) {
		return RenderedInvocation{
			ProviderType: inv.ProviderType,
			AgentID:      inv.Spec.ID,
			SupportLevel: SupportUnsupported,
			Warnings:     targetWarnings(inv),
		}, fmt.Errorf("agentadapter: cursor-cli non listé dans provider_targets")
	}
	prompt := strings.TrimSpace(inv.Prompt)
	if prompt == "" {
		return RenderedInvocation{}, fmt.Errorf("agentadapter: prompt vide")
	}
	cmd := commandOr(inv.Runtime.Command, defaultCursorCommand)
	args := append([]string(nil), inv.Runtime.Args...)
	out := baseRendered(inv, SupportInlinePrompt, cmd, args)
	out.Warnings = append(out.Warnings,
		"profil Cursor natif non installé — utilisation inline_prompt via stdin")
	return out, nil
}
