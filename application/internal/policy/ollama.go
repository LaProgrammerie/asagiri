package policy

import "fmt"

// OllamaAllowedRoles lists permitted Ollama uses (spec §10.1).
var OllamaAllowedRoles = []string{
	"classify_task",
	"detect_risk",
	"select_context_files",
	"summarize_diff",
	"generate_handoff",
	"pre_review",
	"query_rag_index",
}

// OllamaForbiddenRoles lists disallowed Ollama uses by default (spec §10.2).
var OllamaForbiddenRoles = []string{
	"modify_critical_code",
	"sole_validation",
	"decide_db_migration",
	"change_dependencies",
	"modify_secrets",
	"publish_pr_without_external_validation",
}

// CheckOllamaRole returns an error if role is forbidden for Ollama.
func CheckOllamaRole(role string) error {
	for _, forbidden := range OllamaForbiddenRoles {
		if forbidden == role {
			return fmt.Errorf("ollama: rôle %q interdit par défaut (spec §10.2)", role)
		}
	}
	return nil
}

// IsOllamaAgent reports whether the configured agent name refers to Ollama.
func IsOllamaAgent(name string) bool {
	return name == "ollama"
}
