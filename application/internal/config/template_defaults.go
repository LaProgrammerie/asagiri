package config

import "strings"

// Default agent names are logical ids in config.agents (not provider ids).
// Onboarding, CLI flag defaults, and empty-config fallbacks use these keys for
// work.default_* until the operator defines their own profiles.
// ApplyRecommendedRuntimeCatalog seeds matching providers/agents for new configs.
// Legacy configs may still use tool-named agents (kiro, cursor, ollama, …).
const (
	DefaultAgentSpec     = "laprogrammerie"
	DefaultAgentDev      = "dev"
	DefaultAgentReviewer = "reviewer"
	DefaultAgentEnrich   = "local-rag"
)

func workAgentOr(configured, fallback string) string {
	if v := strings.TrimSpace(configured); v != "" {
		return v
	}
	return fallback
}

// WorkSpecAgent returns work.default_spec_agent or the template default logical agent id.
func (c *Config) WorkSpecAgent() string {
	if c == nil {
		return DefaultAgentSpec
	}
	return workAgentOr(c.Work.DefaultSpecAgent, DefaultAgentSpec)
}

// WorkDevAgent returns work.default_agent or the template default logical agent id.
func (c *Config) WorkDevAgent() string {
	if c == nil {
		return DefaultAgentDev
	}
	return workAgentOr(c.Work.DefaultAgent, DefaultAgentDev)
}

// WorkReviewerAgent returns work.default_reviewer or the template default logical agent id.
func (c *Config) WorkReviewerAgent() string {
	if c == nil {
		return DefaultAgentReviewer
	}
	return workAgentOr(c.Work.DefaultReviewer, DefaultAgentReviewer)
}

// WorkEnricherAgent returns work.default_enricher or the template default logical agent id.
func (c *Config) WorkEnricherAgent() string {
	if c == nil {
		return DefaultAgentEnrich
	}
	return workAgentOr(c.Work.DefaultEnricher, DefaultAgentEnrich)
}

// DefaultPremiumReferenceModel is intentionally empty: no baseline invented by default.
// Users must explicitly set pricing.premium_reference_model in config.yaml to enable savings.
const DefaultPremiumReferenceModel = ""


// IsTemplateDefaultProjectName reports whether the project name is unset or still the template default.
func IsTemplateDefaultProjectName(name string) bool {
	n := strings.TrimSpace(name)
	return n == "" || n == "my-project"
}

// IsTemplateDefaultBranchPrefix reports whether worktrees.branch_prefix is unset or still the template default.
func IsTemplateDefaultBranchPrefix(prefix string) bool {
	p := strings.TrimSpace(prefix)
	return p == "" || p == DefaultBranchPrefix
}

// IsTemplateDefaultValidationCommands reports whether validation.commands is empty or matches Go template defaults.
func IsTemplateDefaultValidationCommands(cmds []ValidationCommand) bool {
	if len(cmds) == 0 {
		return true
	}
	defaults := DefaultGoValidationCommands("")
	if len(cmds) != len(defaults) {
		return false
	}
	for i, cmd := range cmds {
		if cmd.Command != defaults[i].Command || cmd.Name != defaults[i].Name {
			return false
		}
	}
	return true
}
