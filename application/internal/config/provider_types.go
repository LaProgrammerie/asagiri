package config

import "strings"

// Provider runtime types — adapter selection uses provider.type only, never the
// provider or agent map key (see agent factory Phase 2).
const (
	ProviderTypeExec       = "exec"
	ProviderTypeClaudeCode = "claude-code"
	ProviderTypeKiroCLI    = "kiro-cli"
	ProviderTypeCursorCLI  = "cursor-cli"
	ProviderTypeCodexCLI   = "codex-cli"
	ProviderTypeGeminiCLI  = "gemini-cli"
	ProviderTypeOllama     = "ollama"
)

// KnownProviderTypes lists provider.type values recognised at config load time.
func KnownProviderTypes() []string {
	return []string{
		ProviderTypeExec,
		ProviderTypeClaudeCode,
		ProviderTypeKiroCLI,
		ProviderTypeCursorCLI,
		ProviderTypeCodexCLI,
		ProviderTypeGeminiCLI,
		ProviderTypeOllama,
	}
}

// IsKnownProviderType reports whether t is a supported provider.type value.
func IsKnownProviderType(t string) bool {
	switch strings.TrimSpace(t) {
	case ProviderTypeExec,
		ProviderTypeClaudeCode,
		ProviderTypeKiroCLI,
		ProviderTypeCursorCLI,
		ProviderTypeCodexCLI,
		ProviderTypeGeminiCLI,
		ProviderTypeOllama:
		return true
	default:
		return false
	}
}
