package intent

import "testing"

// Golden cases lock V1 scope classification for asa work Product Layer routing.
// Priority: technical_task > feature_work > product_level_intent.
// When uncertain, feature_work is expected (Product Layer must not run).
func TestClassifyIntentScopeGolden(t *testing.T) {
	tests := []struct {
		name        string
		instruction string
		want        IntentScope
	}{
		// --- technical_task (highest priority) ---
		{name: "bug fix fr", instruction: "corrige le bug login", want: ScopeTechnicalTask},
		{name: "bug fix en", instruction: "fix bug in CRM login flow", want: ScopeTechnicalTask},
		{name: "refactor", instruction: "refactor parser yaml", want: ScopeTechnicalTask},
		{name: "refactor crm ui", instruction: "refactor CRM dashboard widgets", want: ScopeTechnicalTask},
		{name: "ci lint", instruction: "fix CI lint failures", want: ScopeTechnicalTask},
		{name: "repair error", instruction: "répare l'erreur 500 sur /api/users", want: ScopeTechnicalTask},
		{name: "unit tests beat feature", instruction: "ajoute des tests unitaires pour billing", want: ScopeTechnicalTask},
		{name: "test for crm", instruction: "Créer un test pour le CRM", want: ScopeTechnicalTask},
		{name: "e2e test greenfield wording", instruction: "Créer un test E2E pour checkout", want: ScopeTechnicalTask},
		{name: "parser error", instruction: "parser JSON invalide dans config", want: ScopeTechnicalTask},

		// --- feature_work (default path; beats product when ambiguous) ---
		{name: "add export", instruction: "ajoute export CSV", want: ScopeFeatureWork},
		{name: "implement endpoint", instruction: "implémente endpoint billing", want: ScopeFeatureWork},
		{name: "endpoint on crm", instruction: "Ajouter un endpoint CRM", want: ScopeFeatureWork},
		{name: "implement webhook", instruction: "implement Stripe webhook handler", want: ScopeFeatureWork},
		{name: "auth middleware", instruction: "add auth middleware for API", want: ScopeFeatureWork},
		{name: "release notes", instruction: "latest release notes", want: ScopeFeatureWork},
		{name: "dashboard update", instruction: "mise à jour du dashboard admin", want: ScopeFeatureWork},
		{name: "oauth integration", instruction: "intégrer OAuth Google", want: ScopeFeatureWork},
		{name: "profile page", instruction: "ajoute page profil utilisateur", want: ScopeFeatureWork},
		{name: "module export not product", instruction: "Créer un module export PDF", want: ScopeFeatureWork},
		{name: "rest endpoint on crm", instruction: "Créer un endpoint REST pour le CRM", want: ScopeFeatureWork},
		{name: "empty instruction", instruction: "   ", want: ScopeFeatureWork},

		// --- product_level_intent (narrow greenfield signals only) ---
		{name: "crm artisans", instruction: "Créer un CRM pour artisans", want: ScopeProductLevel},
		{name: "saas billing", instruction: "créer un SaaS de facturation", want: ScopeProductLevel},
		{name: "marketplace b2b", instruction: "construire une marketplace B2B", want: ScopeProductLevel},
		{name: "booking app", instruction: "Créer une app de réservation", want: ScopeProductLevel},
		{name: "training platform", instruction: "concevoir une plateforme de formation", want: ScopeProductLevel},
		{name: "saas for smb", instruction: "Créer un produit SaaS pour PME", want: ScopeProductLevel},
		{name: "multi-tenant crm", instruction: "construire un CRM multi-tenant", want: ScopeProductLevel},
		{name: "saas architecture intent", instruction: "concevoir l'architecture du SaaS", want: ScopeProductLevel},
		{name: "crm token alone", instruction: "développer le CRM existant", want: ScopeProductLevel},
		{name: "improve crm", instruction: "améliorer le CRM clients", want: ScopeProductLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyIntentScope(tt.instruction)
			if got != tt.want {
				t.Fatalf("ClassifyIntentScope(%q) = %q, want %q", tt.instruction, got, tt.want)
			}
		})
	}
}
