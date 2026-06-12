package product

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntentSlugExamples(t *testing.T) {
	tests := []struct {
		intent string
		want   string
	}{
		{"Créer un CRM pour artisans", "crm-artisans"},
		{"Créer une plateforme de réservation pour vétérinaires", "reservation-veterinaires"},
		{"Créer un SaaS de facturation", "saas-facturation"},
	}
	for _, tt := range tests {
		got := IntentSlug(tt.intent)
		if got != tt.want {
			t.Errorf("IntentSlug(%q) = %q, want %q", tt.intent, got, tt.want)
		}
	}
}

func TestIntentSlugFallbackPreservesLegacyShape(t *testing.T) {
	intent := "my custom product name"
	got := IntentSlug(intent)
	want := Slug(intent)
	if got != want {
		t.Fatalf("fallback slug = %q, want %q", got, want)
	}
}

func TestResolveProductIDPrefersExistingLegacyDir(t *testing.T) {
	root := t.TempDir()
	legacy := Slug("Créer un CRM pour artisans")
	legacyDir := filepath.Join(root, ".asagiri", "products", legacy)
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := ResolveProductID(root, "Créer un CRM pour artisans", "")
	if got != legacy {
		t.Fatalf("ResolveProductID = %q, want legacy %q", got, legacy)
	}
}

func TestResolveProductIDUsesDerivedForNewProducts(t *testing.T) {
	root := t.TempDir()
	got := ResolveProductID(root, "Créer un CRM pour artisans", "")
	if got != "crm-artisans" {
		t.Fatalf("ResolveProductID = %q, want crm-artisans", got)
	}
}

func TestResolveProductIDExplicitFeature(t *testing.T) {
	root := t.TempDir()
	got := ResolveProductID(root, "Créer un CRM pour artisans", "my-existing-feature")
	if got != "my-existing-feature" {
		t.Fatalf("explicit feature not honored: %q", got)
	}
}
