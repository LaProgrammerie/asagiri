package config

import "strings"

// Default agent names — single source of truth for fallback values throughout
// the codebase. All packages should reference these constants instead of
// hard-coding agent name strings.
const (
	DefaultAgentSpec     = "kiro"
	DefaultAgentDev      = "cursor"
	DefaultAgentReviewer = "codex"
	DefaultAgentEnrich   = "ollama"
)


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
