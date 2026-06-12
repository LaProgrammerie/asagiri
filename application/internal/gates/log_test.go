package gates

import (
	"encoding/json"
	"testing"
)

func TestNewLogDocumentRunScope(t *testing.T) {
	r := Result{
		GateID:     "plan_gate",
		Status:     VerdictPass,
		Confidence: 0.9,
		Notes:      []string{"ok"},
	}
	doc := NewLogDocument("run-1", "run", "plan", "feat", "reviewer", r, "2026-06-06T12:00:00Z")
	if doc.RunID != "run-1" || doc.ScopeKind != "run" || doc.GateName != "plan" {
		t.Fatalf("doc: %+v", doc)
	}
	if doc.Status != "pass" || doc.Agent != "reviewer" {
		t.Fatalf("status/agent: %+v", doc)
	}
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var back LogDocument
	if err := json.Unmarshal(body, &back); err != nil {
		t.Fatal(err)
	}
	if back.TaskID != "" || back.RunID != "run-1" {
		t.Fatalf("round-trip: %+v", back)
	}
}

func TestNewLogDocumentTaskScope(t *testing.T) {
	r := Result{Status: VerdictWarn, Confidence: 0.5}
	doc := NewLogDocument("task-9", "task", "governance", "feat", "reviewer", r, "2026-06-06T12:00:00Z")
	if doc.TaskID != "task-9" || doc.RunID != "" {
		t.Fatalf("task scope: %+v", doc)
	}
}
