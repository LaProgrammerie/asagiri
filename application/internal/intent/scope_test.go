package intent

import "testing"

func TestClassifyIntentScope(t *testing.T) {
	tests := []struct {
		instruction string
		want        IntentScope
	}{
		{"corrige le bug login", ScopeTechnicalTask},
		{"ajoute export CSV", ScopeFeatureWork},
		{"Créer un CRM pour artisans", ScopeProductLevel},
		{"créer un SaaS de facturation", ScopeProductLevel},
		{"implémente endpoint billing", ScopeFeatureWork},
		{"refactor parser yaml", ScopeTechnicalTask},
		{"Créer un test pour le CRM", ScopeTechnicalTask},
		{"Ajouter un endpoint CRM", ScopeFeatureWork},
		{"latest release notes", ScopeFeatureWork},
	}
	for _, tt := range tests {
		got := ClassifyIntentScope(tt.instruction)
		if got != tt.want {
			t.Errorf("ClassifyIntentScope(%q) = %q, want %q", tt.instruction, got, tt.want)
		}
	}
}

func TestShouldRunProductLayer(t *testing.T) {
	if ShouldRunProductLayer(ScopeProductLevel) != true {
		t.Fatal("expected product layer for product_level_intent")
	}
	if ShouldRunProductLayer(ScopeTechnicalTask) {
		t.Fatal("technical task must not trigger product layer")
	}
	if ShouldRunProductLayer(ScopeFeatureWork) {
		t.Fatal("feature work must not trigger product layer")
	}
}
