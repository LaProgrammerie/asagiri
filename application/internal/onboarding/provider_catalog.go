package onboarding

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// ProviderPreset describes one selectable external runtime in the onboarding wizard.
type ProviderPreset struct {
	ID      string
	Label   string
	Type    string
	Command string
	Args    []string
	Timeout int
}

// ProviderCatalog lists runtimes offered during onboarding step "providers".
func ProviderCatalog() []ProviderPreset {
	return []ProviderPreset{
		{ID: "kiro-cli", Label: "Kiro CLI", Type: config.ProviderTypeKiroCLI, Command: "kiro", Args: []string{"--cli"}},
		{ID: "cursor-cli", Label: "Cursor Agent", Type: config.ProviderTypeCursorCLI, Command: "cursor-agent", Timeout: 3600},
		{ID: "claude-code", Label: "Claude Code", Type: config.ProviderTypeClaudeCode, Command: "claude", Args: []string{"--print", "--output-format", "stream-json"}, Timeout: 600},
		{ID: "codex-cli", Label: "Codex CLI", Type: config.ProviderTypeCodexCLI, Command: "codex", Timeout: 3600},
		{ID: "ollama", Label: "Ollama", Type: config.ProviderTypeOllama, Command: "ollama", Args: []string{"run", "qwen2.5-coder:14b"}, Timeout: 300},
	}
}

// DefaultEnabledProviders returns the default provider ids when onboarding does not override them.
func DefaultEnabledProviders() []string {
	return []string{"kiro-cli", "cursor-cli", "claude-code", "ollama"}
}

// DefaultLogicalAgentNames are examples shown when creating logical agents (step agents).
func DefaultLogicalAgentNames() []string {
	return []string{"laprogrammerie", "dev", "architect", "reviewer", "local-rag"}
}

func parseEnabledProvidersCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func formatEnabledProvidersCSV(ids []string) string {
	return strings.Join(ids, ", ")
}
