package product

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Intent slug rules (V1):
// 1. Strip leading creation verbs ("créer un", "construire une", …).
// 2. Prefer structured joins:
//    - "plateforme de X pour Y" → x-y
//    - "N pour Y"              → n-y   (ex. crm pour artisans)
//    - "N de Y"                → n-y   (ex. saas de facturation)
// 3. Fallback: Slug(raw intent) — preserves legacy slugs when no rule matches.
// 4. ResolveProductID keeps an existing legacy directory if present on disk.

var accentReplacer = strings.NewReplacer(
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"à", "a", "â", "a", "ä", "a",
	"ù", "u", "û", "u", "ü", "u",
	"ô", "o", "ö", "o",
	"î", "i", "ï", "i",
	"ç", "c",
)

var creationPrefixes = []string{
	"créer un ", "creer un ", "créer une ", "creer une ",
	"construire un ", "construire une ",
	"concevoir un ", "concevoir une ",
	"developper un ", "développer un ", "developper une ", "développer une ",
}

var slugFillerTokens = map[string]bool{
	"un": true, "une": true, "le": true, "la": true, "les": true,
	"du": true, "de": true, "des": true, "pour": true, "d": true,
}

// IntentSlug derives a compact business slug from a natural-language product intent.
func IntentSlug(intent string) string {
	if derived := deriveBusinessSlug(intent); derived != "" {
		return derived
	}
	return Slug(intent)
}

// ResolveProductID picks a product slug without breaking existing on-disk products.
func ResolveProductID(repoRoot, instruction, explicitFeature string) string {
	if explicitFeature != "" {
		s := Slug(explicitFeature)
		if s != "" && s != "product" {
			return s
		}
	}

	derived := IntentSlug(instruction)
	legacy := Slug(instruction)
	if derived == legacy {
		return derived
	}
	if productDirExists(repoRoot, legacy) && !productDirExists(repoRoot, derived) {
		return legacy
	}
	return derived
}

func deriveBusinessSlug(intent string) string {
	s := normalizeIntentText(intent)
	if s == "" {
		return ""
	}
	s = stripCreationPrefix(s)

	if rest, ok := strings.CutPrefix(s, "plateforme de "); ok {
		if before, after, found := strings.Cut(rest, " pour "); found {
			return joinSlugParts(before, after)
		}
	}

	if before, after, found := strings.Cut(s, " pour "); found {
		return joinSlugParts(before, after)
	}

	if before, after, found := strings.Cut(s, " de "); found {
		return joinSlugParts(before, after)
	}

	cleaned := strings.Join(filterSlugTokens(scopeTokens(s)), "-")
	if cleaned != "" && cleaned != "product" {
		return cleaned
	}
	return ""
}

func normalizeIntentText(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = accentReplacer.Replace(s)
	return strings.TrimSpace(s)
}

func stripCreationPrefix(s string) string {
	for _, prefix := range creationPrefixes {
		if strings.HasPrefix(s, prefix) {
			return strings.TrimSpace(s[len(prefix):])
		}
	}
	return s
}

func joinSlugParts(parts ...string) string {
	var tokens []string
	for _, part := range parts {
		tokens = append(tokens, filterSlugTokens(scopeTokens(part))...)
	}
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, "-")
}

func filterSlugTokens(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if tok == "" || slugFillerTokens[tok] {
			continue
		}
		out = append(out, tok)
	}
	return out
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

func productDirExists(repoRoot, productID string) bool {
	path := filepath.Join(repoRoot, ".asagiri", "products", Slug(productID))
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
