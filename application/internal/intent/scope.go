package intent

import (
	"strings"
	"unicode"
)

// IntentScope classifies how broad a work instruction is (product layer V1).
type IntentScope string

const (
	ScopeTechnicalTask IntentScope = "technical_task"
	ScopeFeatureWork   IntentScope = "feature_work"
	ScopeProductLevel  IntentScope = "product_level_intent"
)

var (
	technicalTokens = map[string]bool{
		"bug": true, "fix": true, "test": true, "tests": true, "refactor": true,
		"parser": true, "erreur": true, "ci": true, "lint": true,
		"corrige": true, "corriger": true, "répare": true, "repare": true,
	}
	featureTokens = map[string]bool{
		"ajoute": true, "ajouter": true, "implémente": true, "implemente": true,
		"implement": true, "endpoint": true, "page": true, "export": true,
		"auth": true, "billing": true,
	}
	productTokens = map[string]bool{
		"crm": true, "saas": true, "marketplace": true, "plateforme": true, "produit": true,
	}
	productPhrases = []string{
		"créer un saas", "creer un saas", "créer une app", "creer une app",
		"construire une marketplace", "application complète", "application complete",
	}
)

// ClassifyIntentScope returns a deterministic scope from the raw instruction.
// Technical and feature signals take priority over product signals.
// When uncertain, returns ScopeFeatureWork (product layer is not triggered).
// Golden cases: scope_golden_test.go (V1 contract).
func ClassifyIntentScope(instruction string) IntentScope {
	lower := normalizeScopeText(instruction)
	if lower == "" {
		return ScopeFeatureWork
	}
	tokens := scopeTokenSet(lower)

	for tok := range tokens {
		if technicalTokens[tok] {
			return ScopeTechnicalTask
		}
	}

	for tok := range tokens {
		if featureTokens[tok] {
			return ScopeFeatureWork
		}
	}

	for _, phrase := range productPhrases {
		if strings.Contains(lower, phrase) {
			return ScopeProductLevel
		}
	}

	for tok := range tokens {
		if productTokens[tok] {
			return ScopeProductLevel
		}
	}

	if hasCreationPrefix(lower) && looksLikeGreenfieldProduct(lower) {
		return ScopeProductLevel
	}

	return ScopeFeatureWork
}

// ShouldRunProductLayer reports whether the product preparation pipeline applies.
func ShouldRunProductLayer(scope IntentScope) bool {
	return scope == ScopeProductLevel
}

func hasCreationPrefix(lower string) bool {
	prefixes := []string{
		"créer un ", "creer un ", "créer une ", "creer une ",
		"construire un ", "construire une ", "concevoir ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func looksLikeGreenfieldProduct(lower string) bool {
	for tok := range scopeTokenSet(lower) {
		if productTokens[tok] {
			return true
		}
	}
	return strings.Contains(lower, "application") || strings.Contains(lower, "plateforme")
}

func normalizeScopeText(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func scopeTokenSet(lower string) map[string]bool {
	set := make(map[string]bool)
	for _, tok := range scopeTokens(lower) {
		set[tok] = true
	}
	return set
}

func scopeTokens(s string) []string {
	s = strings.NewReplacer(",", " ", ".", " ", ":", " ", ";", " ", "'", " ", "\"", " ").Replace(s)
	fields := strings.Fields(s)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.Trim(f, "-")
		if f == "" {
			continue
		}
		var b strings.Builder
		for _, r := range f {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
				b.WriteRune(r)
			}
		}
		if tok := b.String(); tok != "" {
			out = append(out, tok)
		}
	}
	return out
}
