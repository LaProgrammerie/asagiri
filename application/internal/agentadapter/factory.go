package agentadapter

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

var typedAdapters = []Adapter{
	cursorAdapter{},
	kiroAdapter{},
	claudeCodeAdapter{},
	codexAdapter{},
}

// Factory selects adapters by config provider.type (render-only).
type Factory struct {
	fallback Adapter
}

// NewFactory returns a factory with inline fallback for unknown exec-like providers.
func NewFactory() *Factory {
	return &Factory{fallback: inlineAdapter{}}
}

// AdapterFor returns the adapter for providerType or nil.
func AdapterFor(providerType string) Adapter {
	f := NewFactory()
	return f.adapterFor(providerType)
}

// Explain returns human-readable adapter documentation for providerType.
func Explain(providerType string) string {
	if a := AdapterFor(providerType); a != nil {
		return a.Explain()
	}
	return inlineAdapter{}.Explain()
}

func (f *Factory) adapterFor(providerType string) Adapter {
	pt := strings.TrimSpace(providerType)
	for _, a := range typedAdapters {
		if a.Supports(pt, agentspec.Spec{}) {
			return a
		}
	}
	if f.fallback.Supports(pt, agentspec.Spec{}) {
		return f.fallback
	}
	return nil
}

// NewInvocation builds an Invocation from AgentSpec, context and merged runtime config.
func NewInvocation(providerType, agentConfigKey string, spec agentspec.Spec, ctx agentcontext.ExecutionContext, runtime config.Agent) Invocation {
	prompt := agentcontext.RenderPrompt(ctx)
	return Invocation{
		ProviderType:   strings.TrimSpace(providerType),
		AgentConfigKey: strings.TrimSpace(agentConfigKey),
		Spec:           spec,
		Context:        ctx,
		Prompt:         prompt,
		Runtime:        runtime,
	}
}

// Render produces a RenderedInvocation using the adapter for inv.ProviderType.
func (f *Factory) Render(inv Invocation) (RenderedInvocation, error) {
	if strings.TrimSpace(inv.Prompt) == "" {
		return RenderedInvocation{}, fmt.Errorf("agentadapter: prompt vide")
	}
	a := f.adapterFor(inv.ProviderType)
	if a == nil {
		return f.renderUnsupported(inv)
	}
	if !a.Supports(inv.ProviderType, inv.Spec) {
		return f.renderWithInlineFallback(inv, []string{
			fmt.Sprintf("adapter %T décliné pour provider.type %q (provider_targets)", a, inv.ProviderType),
		})
	}
	out, err := a.Render(inv)
	if err != nil {
		return f.renderWithInlineFallback(inv, []string{err.Error()})
	}
	return out, nil
}

// RenderFromConfig resolves provider.type from config and renders the invocation.
func RenderFromConfig(cfg *config.Config, agentConfigKey string, spec agentspec.Spec, ctx agentcontext.ExecutionContext) (RenderedInvocation, error) {
	if cfg == nil {
		return RenderedInvocation{}, fmt.Errorf("agentadapter: config nil")
	}
	providerType, merged, err := cfg.MergedAgentRuntime(agentConfigKey)
	if err != nil {
		return RenderedInvocation{}, err
	}
	inv := NewInvocation(providerType, agentConfigKey, spec, ctx, merged)
	return NewFactory().Render(inv)
}

func (f *Factory) renderUnsupported(inv Invocation) (RenderedInvocation, error) {
	warnings := []string{
		fmt.Sprintf("provider.type %q sans adapter dédié", inv.ProviderType),
	}
	if f.fallback.Supports(inv.ProviderType, inv.Spec) {
		return f.renderWithInlineFallback(inv, warnings)
	}
	return RenderedInvocation{
		ProviderType: inv.ProviderType,
		AgentID:      inv.Spec.ID,
		SupportLevel: SupportUnsupported,
		Warnings:     warnings,
	}, fmt.Errorf("agentadapter: provider.type %q non supporté", inv.ProviderType)
}

func (f *Factory) renderWithInlineFallback(inv Invocation, extraWarnings []string) (RenderedInvocation, error) {
	out, err := f.fallback.Render(inv)
	if err != nil {
		out.SupportLevel = SupportUnsupported
		out.Warnings = append(extraWarnings, out.Warnings...)
		return out, err
	}
	out.Warnings = append(extraWarnings, out.Warnings...)
	if out.SupportLevel == SupportInlinePrompt {
		out.Warnings = append(out.Warnings, "invocation produite via fallback inline générique")
	}
	return out, nil
}

// SupportMatrix returns provider.type → support level for a spec (dry analysis).
func SupportMatrix(spec agentspec.Spec) map[string]SupportLevel {
	out := make(map[string]SupportLevel)
	for _, pt := range config.KnownProviderTypes() {
		f := NewFactory()
		a := f.adapterFor(pt)
		if a == nil {
			out[pt] = SupportUnsupported
			continue
		}
		if !a.Supports(pt, spec) {
			out[pt] = SupportUnsupported
			continue
		}
		switch a.(type) {
		case claudeCodeAdapter:
			out[pt] = SupportNativeProfile
		default:
			out[pt] = SupportInlinePrompt
		}
	}
	return out
}
