package config

import (
	"fmt"
	"strings"
)

// ProviderConfig describes how Asagiri talks to an external runtime tool.
type ProviderConfig struct {
	Type    string            `yaml:"type"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env,omitempty"`
	Timeout int               `yaml:"timeout,omitempty"`
}

// LookupProvider returns a declared providers: entry by name.
func (c *Config) LookupProvider(name string) (ProviderConfig, error) {
	if c == nil {
		return ProviderConfig{}, fmt.Errorf("config nil")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ProviderConfig{}, fmt.Errorf("provider: nom requis")
	}
	if c.Providers == nil {
		return ProviderConfig{}, fmt.Errorf("provider %q inconnu (section providers absente ou vide)", name)
	}
	p, ok := c.Providers[name]
	if !ok {
		return ProviderConfig{}, fmt.Errorf("provider %q inconnu (absent de config.providers)", name)
	}
	return p, nil
}

// LookupAgent returns a declared agents: entry by name.
func (c *Config) LookupAgent(name string) (Agent, error) {
	if c == nil {
		return Agent{}, fmt.Errorf("config nil")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Agent{}, fmt.Errorf("agent: nom requis")
	}
	if c.Agents == nil {
		return Agent{}, fmt.Errorf("agent %q inconnu (section agents absente ou vide)", name)
	}
	a, ok := c.Agents[name]
	if !ok {
		return Agent{}, fmt.Errorf("agent %q inconnu (absent de config.agents)", name)
	}
	return a, nil
}

// AgentProviderType returns the provider.type for a named agent.
// Legacy agents without provider field implicitly use exec.
func (c *Config) AgentProviderType(agentName string) (string, error) {
	a, err := c.LookupAgent(agentName)
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(a.Provider)
	if ref == "" {
		return ProviderTypeExec, nil
	}
	p, err := c.LookupProvider(ref)
	if err != nil {
		return "", err
	}
	typ := strings.TrimSpace(p.Type)
	if typ == "" {
		return "", fmt.Errorf("agents.%s.provider %q: type manquant sur le provider", agentName, ref)
	}
	return typ, nil
}

func (c *Config) validateProvidersAndAgents() error {
	if c.Providers == nil {
		c.Providers = map[string]ProviderConfig{}
	}
	for name, p := range c.Providers {
		typ := strings.TrimSpace(p.Type)
		if typ == "" {
			return fmt.Errorf("providers.%s.type: requis", name)
		}
		if !IsKnownProviderType(typ) {
			return fmt.Errorf("providers.%s.type: %q inconnu (types supportés: %s)",
				name, typ, strings.Join(KnownProviderTypes(), ", "))
		}
	}
	for agentName, a := range c.Agents {
		ref := strings.TrimSpace(a.Provider)
		if ref == "" {
			continue
		}
		if _, err := c.LookupProvider(ref); err != nil {
			return fmt.Errorf("agents.%s.provider: %w", agentName, err)
		}
	}
	return nil
}

// MergedAgentRuntime resolves provider.type and the merged runtime Agent config
// (provider defaults + agent overrides) for factory construction.
func (c *Config) MergedAgentRuntime(agentName string) (providerType string, merged Agent, err error) {
	a, err := c.LookupAgent(agentName)
	if err != nil {
		return "", Agent{}, err
	}
	providerType, err = c.AgentProviderType(agentName)
	if err != nil {
		return "", Agent{}, err
	}
	merged = a
	ref := strings.TrimSpace(a.Provider)
	if ref == "" {
		return providerType, merged, nil
	}
	p, err := c.LookupProvider(ref)
	if err != nil {
		return "", Agent{}, err
	}
	return providerType, mergeAgentWithProvider(p, a), nil
}

func mergeAgentWithProvider(p ProviderConfig, a Agent) Agent {
	merged := a
	if strings.TrimSpace(merged.Command) == "" {
		merged.Command = p.Command
	}
	merged.Args = append(append([]string{}, p.Args...), a.Args...)
	merged.Env = mergeStringMaps(p.Env, a.Env)
	if merged.Timeout == 0 {
		merged.Timeout = p.Timeout
	}
	return merged
}

func mergeStringMaps(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}
