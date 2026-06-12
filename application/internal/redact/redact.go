package redact

import (
	"regexp"
	"strings"
)

var (
	reBearer = regexp.MustCompile(`(?i)(bearer\s+)[a-z0-9._\-]+`)
	reSecret = regexp.MustCompile(`(?i)(secret_[a-z0-9]+|sk-[a-z0-9]+|ghp_[a-z0-9]+|xox[baprs]-[a-z0-9\-]+)`)
	reKV     = regexp.MustCompile(`(?i)(api[_-]?key|token|password|authorization)\s*[:=]\s*\S+`)
)

// String masks common secret patterns in log or error output.
func String(s string) string {
	if s == "" {
		return s
	}
	out := reBearer.ReplaceAllString(s, `${1}[REDACTED]`)
	out = reSecret.ReplaceAllString(out, "[REDACTED]")
	out = reKV.ReplaceAllString(out, `$1=[REDACTED]`)
	return out
}

// EnvValue returns a safe representation of an env var value for logs.
func EnvValue(key, value string) string {
	kl := strings.ToLower(key)
	if strings.Contains(kl, "token") || strings.Contains(kl, "secret") || strings.Contains(kl, "password") || strings.Contains(kl, "key") {
		if value == "" {
			return ""
		}
		return "[REDACTED]"
	}
	return value
}
