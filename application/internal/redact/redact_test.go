package redact

import (
	"strings"
	"testing"
)

func TestStringRedactsSecrets(t *testing.T) {
	in := "NOTION_TOKEN=secret_abc123 and Bearer eyJhbGciOiJIUzI1NiJ9"
	out := String(in)
	if out == in {
		t.Fatalf("expected redaction, got %q", out)
	}
	if strings.Contains(out, "secret_abc123") {
		t.Fatalf("leaked secret: %q", out)
	}
}
