package onboarding

import "strings"

// validationName extracts a short validation identifier from a preview string.
func validationName(raw string) string {
	raw = strings.TrimSpace(raw)
	if i := strings.IndexAny(raw, ":("); i > 0 {
		raw = strings.TrimSpace(raw[:i])
	}
	if raw == "" {
		return "validation"
	}
	return raw
}
