package gates

import (
	"fmt"
	"strings"
)

// Summary returns a short human-readable reason for the result.
func Summary(r Result) string {
	if r.ParseError != "" {
		return r.ParseError
	}
	if len(r.Notes) > 0 {
		return r.Notes[0]
	}
	if len(r.Findings) > 0 {
		return r.Findings[0].Message
	}
	return "gate failed"
}

// FormatFailure builds an actionable failure message from notes and findings.
func FormatFailure(r Result) string {
	var parts []string
	if msg := Summary(r); msg != "" {
		parts = append(parts, msg)
	}
	for _, f := range r.Findings {
		line := fmt.Sprintf("[%s/%s] %s", f.Code, f.Severity, f.Message)
		if len(f.Actions) > 0 {
			line += " — actions: " + strings.Join(f.Actions, "; ")
		}
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return "gate failed"
	}
	return strings.Join(parts, " | ")
}
