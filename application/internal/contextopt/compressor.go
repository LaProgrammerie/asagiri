package contextopt

import (
	"strings"
)

// CompressSections truncates markdown-style sections while keeping headings (specv3 V3.1 — no LLM).
func CompressSections(markown string, maxSectionChars int) string {
	if maxSectionChars <= 0 {
		maxSectionChars = 8000
	}
	var b strings.Builder
	parts := strings.Split(markown, "\n##")
	for i, p := range parts {
		if i == 0 {
			b.WriteString(trimSection(p, maxSectionChars))
			continue
		}
		b.WriteString("\n##")
		b.WriteString(trimSection(p, maxSectionChars))
	}
	return b.String()
}

func trimSection(section string, limit int) string {
	if len(section) <= limit {
		return section
	}
	return section[:limit] + "\n… section trimmed …\n"
}
