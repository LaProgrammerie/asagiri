package workcli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/intent"
)

func TestHandleProductLayerUsesDerivedSlug(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer
	out, err := HandleProductLayer(&buf, "Créer un CRM pour artisans", intent.ScopeProductLevel, intent.ResolvedIntent{
		Action: intent.IntentDevelop,
	}, intent.WorkOptions{DryRun: true}, root)
	if err != nil {
		t.Fatal(err)
	}
	if out.ProductID != "crm-artisans" {
		t.Fatalf("product id = %q, want crm-artisans", out.ProductID)
	}
	if !strings.Contains(buf.String(), "Product ID: crm-artisans") {
		t.Fatalf("output should show derived slug: %s", buf.String())
	}
}

func TestHandleProductLayerPlanOnlyMessage(t *testing.T) {
	root := t.TempDir()
	var buf bytes.Buffer
	_, err := HandleProductLayer(&buf, "Créer un CRM pour artisans", intent.ScopeProductLevel, intent.ResolvedIntent{
		Action: intent.IntentDevelop,
	}, intent.WorkOptions{PlanOnly: true}, root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "plan-only") {
		t.Fatalf("missing plan-only note: %s", buf.String())
	}
}

func TestHandleProductLayerNonInteractiveHint(t *testing.T) {
	root := t.TempDir()
	_, err := HandleProductLayer(bytes.NewBuffer(nil), "Créer un CRM pour artisans", intent.ScopeProductLevel, intent.ResolvedIntent{
		Action: intent.IntentDevelop,
	}, intent.WorkOptions{Interactive: false, Yes: false}, root)
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	var confirm *intent.ConfirmationRequiredError
	if !errors.As(err, &confirm) {
		t.Fatalf("expected ConfirmationRequiredError, got %T: %v", err, err)
	}
	if !strings.Contains(confirm.Message, `--yes`) || !strings.Contains(confirm.Message, `--dry-run`) {
		t.Fatalf("unexpected message: %s", confirm.Message)
	}
}

func TestWorkDryRunPaths(t *testing.T) {
	cases := []struct {
		name        string
		instruction string
		wantScope   intent.IntentScope
		wantProduct bool
		wantSlug    string
	}{
		{
			name:        "technical",
			instruction: "corrige le bug login",
			wantScope:   intent.ScopeTechnicalTask,
			wantProduct: false,
		},
		{
			name:        "feature",
			instruction: "ajoute export CSV",
			wantScope:   intent.ScopeFeatureWork,
			wantProduct: false,
		},
		{
			name:        "product",
			instruction: "Créer un CRM pour artisans",
			wantScope:   intent.ScopeProductLevel,
			wantProduct: true,
			wantSlug:    "crm-artisans",
		},
	}

	root := t.TempDir()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scope := intent.ClassifyIntentScope(tc.instruction)
			if scope != tc.wantScope {
				t.Fatalf("scope = %q, want %q", scope, tc.wantScope)
			}

			var buf bytes.Buffer
			out, err := HandleProductLayer(&buf, tc.instruction, scope, intent.ResolvedIntent{
				Action: intent.IntentDevelop,
			}, intent.WorkOptions{DryRun: true}, root)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantProduct {
				if !out.Handled || !out.StopWorkFlow {
					t.Fatalf("expected product layer handling: %+v", out)
				}
				if out.ProductID != tc.wantSlug {
					t.Fatalf("product id = %q, want %q", out.ProductID, tc.wantSlug)
				}
				if !strings.Contains(buf.String(), "Product-level intent detected") {
					t.Fatalf("unexpected output: %s", buf.String())
				}
				return
			}
			if out.Handled || out.StopWorkFlow || buf.Len() > 0 {
				t.Fatalf("should not enter product layer: handled=%v stop=%v out=%q", out.Handled, out.StopWorkFlow, buf.String())
			}
			if intent.ShouldRunProductLayer(scope) {
				t.Fatal("scope should not run product layer")
			}
		})
	}
}

func TestWorkDryRunFunctionalOutput(t *testing.T) {
	cases := []struct {
		name        string
		instruction string
	}{
		{"technical", "corrige le bug login"},
		{"feature", "ajoute export CSV"},
		{"product", "Créer un CRM pour artisans"},
	}
	root := t.TempDir()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scope := intent.ClassifyIntentScope(tc.instruction)
			var buf bytes.Buffer
			out, err := HandleProductLayer(&buf, tc.instruction, scope, intent.ResolvedIntent{
				Action: intent.IntentDevelop,
			}, intent.WorkOptions{DryRun: true}, root)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("instruction=%q scope=%s handled=%v stop=%v product_id=%q\n--- output ---\n%s--- end ---",
				tc.instruction, scope, out.Handled, out.StopWorkFlow, out.ProductID, buf.String())
		})
	}
}
