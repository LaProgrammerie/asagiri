package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestPlainUIFallback(t *testing.T) {
	var buf bytes.Buffer
	ui := NewUI(&config.Config{UI: config.UIConfig{Mode: "plain"}}, &buf, false)
	ui.Box("Estimated execution", "Cost: €0.08\n")
	if !strings.Contains(buf.String(), "Estimated execution") {
		t.Fatalf("plain output: %q", buf.String())
	}
}

func TestJSONUIMode(t *testing.T) {
	var buf bytes.Buffer
	ui := newUIWithMode(ModeJSON, &buf, false)
	ui.Event("estimate", map[string]any{"cost": "0.08"})
	if !strings.Contains(buf.String(), "estimate") {
		t.Fatalf("json output: %q", buf.String())
	}
}
