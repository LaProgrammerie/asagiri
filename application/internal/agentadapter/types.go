package agentadapter

import (
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// SupportLevel describes how the provider consumes the orchestrated AgentSpec.
type SupportLevel string

const (
	// SupportNativeProfile — provider flags structurés (ex. claude --print stream-json) ; prompt via stdin.
	SupportNativeProfile SupportLevel = "native_profile"
	// SupportInlinePrompt — prompt orchestré via stdin (chemin sûr par défaut).
	SupportInlinePrompt SupportLevel = "inline_prompt"
	// SupportUnsupported — provider ou cible incompatible ; pas d'invocation fiable.
	SupportUnsupported SupportLevel = "unsupported"
)

// ProviderTarget names a provider.type compatible with an AgentSpec.
type ProviderTarget struct {
	Type string `json:"type"`
}

// Invocation binds AgentSpec, ExecutionContext, runtime config and rendered prompt.
// No subprocess is started.
type Invocation struct {
	ProviderType   string
	AgentConfigKey string
	Spec           agentspec.Spec
	Context        agentcontext.ExecutionContext
	Prompt         string
	Runtime        config.Agent
}

// RenderedInvocation is a deterministic CLI invocation plan (render-only).
type RenderedInvocation struct {
	ProviderType string            `json:"provider_type"`
	AgentID      string            `json:"agent_id"`
	SupportLevel SupportLevel      `json:"support_level"`
	Command      string            `json:"command"`
	Args         []string          `json:"args,omitempty"`
	StdinPrompt  string            `json:"stdin_prompt"`
	Env          map[string]string `json:"env,omitempty"`
	Warnings     []string          `json:"warnings,omitempty"`
}

// Adapter renders provider-specific invocations without executing them.
type Adapter interface {
	Supports(providerType string, spec agentspec.Spec) bool
	Render(inv Invocation) (RenderedInvocation, error)
	Explain() string
}
