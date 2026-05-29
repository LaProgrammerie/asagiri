package components

import "strings"

// Breadcrumb renders navigation crumbs joined by ›.
func Breadcrumb(parts ...string) string {
	trimmed := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			trimmed = append(trimmed, p)
		}
	}
	if len(trimmed) == 0 {
		return ""
	}
	return strings.Join(trimmed, " › ")
}
