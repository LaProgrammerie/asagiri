package agentadapter

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// inlineAdapter is the generic stdin fallback for exec-like providers.
type inlineAdapter struct{}

func (inlineAdapter) Supports(providerType string, _ agentspec.Spec) bool {
	switch strings.TrimSpace(providerType) {
	case config.ProviderTypeExec,
		config.ProviderTypeOllama,
		config.ProviderTypeGeminiCLI,
		"":
		return true
	default:
		return false
	}
}

func (inlineAdapter) Explain() string {
	return "Fallback générique : commande runtime + prompt orchestré sur stdin (aucun profil provider installé)."
}

func (inlineAdapter) Render(inv Invocation) (RenderedInvocation, error) {
	cmd := strings.TrimSpace(inv.Runtime.Command)
	if cmd == "" {
		return RenderedInvocation{
			ProviderType: inv.ProviderType,
			AgentID:      inv.Spec.ID,
			SupportLevel: SupportUnsupported,
			Warnings:     []string{"commande runtime manquante — impossible de produire une invocation inline"},
		}, fmt.Errorf("agentadapter: commande manquante pour provider %q", inv.ProviderType)
	}
	prompt := strings.TrimSpace(inv.Prompt)
	if prompt == "" {
		return RenderedInvocation{}, fmt.Errorf("agentadapter: prompt vide")
	}
	warnings := targetWarnings(inv)
	return RenderedInvocation{
		ProviderType: inv.ProviderType,
		AgentID:      inv.Spec.ID,
		SupportLevel: supportLevelFor(inv.ProviderType, SupportInlinePrompt),
		Command:      cmd,
		Args:         append([]string(nil), inv.Runtime.Args...),
		StdinPrompt:  prompt,
		Env:          copyEnv(inv.Runtime.Env),
		Warnings:     warnings,
	}, nil
}

func baseRendered(inv Invocation, level SupportLevel, cmd string, args []string) RenderedInvocation {
	return RenderedInvocation{
		ProviderType: inv.ProviderType,
		AgentID:      inv.Spec.ID,
		SupportLevel: level,
		Command:      cmd,
		Args:         args,
		StdinPrompt:  strings.TrimSpace(inv.Prompt),
		Env:          copyEnv(inv.Runtime.Env),
		Warnings:     targetWarnings(inv),
	}
}

func targetWarnings(inv Invocation) []string {
	if len(inv.Spec.ProviderTargets) == 0 {
		return nil
	}
	pt := strings.TrimSpace(inv.ProviderType)
	for _, t := range inv.Spec.ProviderTargets {
		if strings.TrimSpace(t) == pt {
			return nil
		}
	}
	return []string{
		fmt.Sprintf("provider.type %q absent de agentspec.provider_targets %v — invocation inline de repli", pt, inv.Spec.ProviderTargets),
	}
}

func targetCompatible(providerType string, spec agentspec.Spec) bool {
	if len(spec.ProviderTargets) == 0 {
		return true
	}
	pt := strings.TrimSpace(providerType)
	for _, t := range spec.ProviderTargets {
		if strings.TrimSpace(t) == pt {
			return true
		}
	}
	return false
}

func copyEnv(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func commandOr(runtimeCmd, fallback string) string {
	if c := strings.TrimSpace(runtimeCmd); c != "" {
		return c
	}
	return fallback
}

func supportLevelFor(providerType string, preferred SupportLevel) SupportLevel {
	if strings.TrimSpace(providerType) == "" {
		return SupportUnsupported
	}
	return preferred
}
