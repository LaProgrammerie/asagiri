package intent

import (
	"strings"
	"testing"
)

func TestProductLayerNonInteractiveMessage(t *testing.T) {
	msg := ProductLayerNonInteractiveMessage("Créer un CRM pour artisans")
	if !strings.Contains(msg, "Product-level intent detected") {
		t.Fatal("missing header")
	}
	if !strings.Contains(msg, `--yes`) || !strings.Contains(msg, `--dry-run`) {
		t.Fatalf("missing actionable hints: %s", msg)
	}
	if !strings.Contains(msg, "Créer un CRM pour artisans") {
		t.Fatal("instruction not echoed")
	}
}
