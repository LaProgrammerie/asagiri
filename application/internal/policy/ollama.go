package policy

import "fmt"

// roleCanon is the single canonical source of truth for the Ollama role policy.
//
// It binds the allowed and forbidden role sets to the current project canon
// (docs/ai/ — see 01-product.md and 02-architecture.md) instead of a stale,
// historical spec reference (the archived specv3 §10.1/§10.2). Having one place
// that owns both lists removes the double, drifting source of truth flagged by
// AUD-006 and lets the Policy_Coherence_Check verify alignment with the canon.
//
// Invariant: Allowed and Forbidden are disjoint (Allowed ∩ Forbidden = ∅); a
// role is never both permitted and forbidden.
type roleCanon struct {
	// Allowed lists the Ollama uses permitted by the current canon.
	Allowed []string
	// Forbidden lists the Ollama uses disallowed by default by the current canon.
	Forbidden []string
}

// ollamaRoleCanon is the SINGLE canonical source for Ollama roles. The exported
// OllamaAllowedRoles and OllamaForbiddenRoles derive from it, so there is no
// duplicated or divergent definition to keep in sync. The role sets reflect the
// current canon (docs/ai/) and not the archived spec sections.
var ollamaRoleCanon = roleCanon{
	Allowed: []string{
		"classify_task",
		"detect_risk",
		"select_context_files",
		"summarize_diff",
		"generate_handoff",
		"pre_review",
		"query_rag_index",
	},
	Forbidden: []string{
		"modify_critical_code",
		"sole_validation",
		"decide_db_migration",
		"change_dependencies",
		"modify_secrets",
		"publish_pr_without_external_validation",
	},
}

// OllamaAllowedRoles lists permitted Ollama uses, derived from the canonical
// source ollamaRoleCanon (current canon, see docs/ai/).
var OllamaAllowedRoles = ollamaRoleCanon.Allowed

// OllamaForbiddenRoles lists disallowed Ollama uses by default, derived from the
// canonical source ollamaRoleCanon (current canon, see docs/ai/).
var OllamaForbiddenRoles = ollamaRoleCanon.Forbidden

// CheckOllamaRole returns an error if role is forbidden for Ollama. The refusal
// names the offending role and never panics; an allowed or unknown role yields
// a nil error.
func CheckOllamaRole(role string) error {
	for _, forbidden := range OllamaForbiddenRoles {
		if forbidden == role {
			return fmt.Errorf("ollama: rôle %q interdit par défaut (canon courant, docs/ai/)", role)
		}
	}
	return nil
}

// IsOllamaAgent reports whether the configured agent name refers to Ollama.
func IsOllamaAgent(name string) bool {
	return name == "ollama"
}
