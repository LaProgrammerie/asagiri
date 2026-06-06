package config

// ApplyRecommendedRuntimeCatalog seeds providers: and agents: for empty configs
// (tests, onboarding bootstrap). It does not run when either map is non-empty,
// so legacy user configs are never overwritten.
func ApplyRecommendedRuntimeCatalog(c *Config) {
	if c == nil {
		return
	}
	if len(c.Agents) > 0 || len(c.Providers) > 0 {
		return
	}
	if c.Providers == nil {
		c.Providers = map[string]ProviderConfig{}
	}
	if c.Agents == nil {
		c.Agents = map[string]Agent{}
	}

	c.Providers["kiro-cli"] = ProviderConfig{
		Type:    ProviderTypeKiroCLI,
		Command: "kiro",
		Args:    []string{"--cli"},
	}
	c.Providers["claude-code"] = ProviderConfig{
		Type:    ProviderTypeClaudeCode,
		Command: "claude",
		Args:    []string{"--print", "--output-format", "stream-json"},
		Timeout: 600,
	}
	c.Providers["ollama"] = ProviderConfig{
		Type:    ProviderTypeOllama,
		Command: "ollama",
		Args:    []string{"run", "qwen2.5-coder:14b"},
		Timeout: 300,
	}

	c.Agents[DefaultAgentSpec] = Agent{
		Provider: "kiro-cli",
		Profile:  DefaultAgentSpec,
	}
	c.Agents[DefaultAgentDev] = Agent{
		Provider: "kiro-cli",
	}
	c.Agents[DefaultAgentReviewer] = Agent{
		Provider: "claude-code",
	}
	c.Agents[DefaultAgentEnrich] = Agent{
		Provider: "ollama",
		Endpoint: "http://localhost:11434",
		Model:    "qwen2.5-coder:14b",
	}
}

// DefaultLogicalAgentIDs returns the recommended bootstrap agent ids (work.default_*).
func DefaultLogicalAgentIDs() []string {
	return []string{DefaultAgentSpec, DefaultAgentDev, DefaultAgentReviewer, DefaultAgentEnrich}
}
