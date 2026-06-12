package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

func TestGatesMarkdownFromGatesHistory(t *testing.T) {
	payload := `{"title":"t1","gates":{"history":[{"gate":"governance","status":"warn","confidence":0.72,"notes":["API drift"]}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: "task-1", PayloadJSON: payload}})
	for _, want := range []string{"## Gates", "governance", "task-1", "warn", "0.72", "API drift"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownEnrichFromGatesHistory(t *testing.T) {
	payload := `{"gates":{"history":[{"gate":"enrich","status":"pass","confidence":0.95,"notes":["ready for dev"]}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: "task-enrich", PayloadJSON: payload}})
	for _, want := range []string{"## Gates", "enrich", "task-enrich", "pass", "0.95", "ready for dev"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownEnrichGovernanceHumanReviewTogether(t *testing.T) {
	taskID := "task-all-gates"
	payload := `{"gates":{"history":[
		{"gate":"enrich","status":"pass","confidence":0.9,"notes":["enrich ok"]},
		{"gate":"governance","status":"pass","confidence":0.85,"notes":["gov ok"]},
		{"gate":"human_review","status":"warn","confidence":1,"notes":["minor doc gap"]}
	]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: taskID, PayloadJSON: payload}})
	for _, want := range []string{"enrich", "governance", "human_review", taskID, "enrich ok", "gov ok", "minor doc gap"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownVerifyEvidenceFromGatesHistory(t *testing.T) {
	payload := `{"gates":{"history":[{"gate":"verify_evidence","status":"pass","confidence":0.93,"notes":["validation bundle ok"]}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: "task-ve-payload", PayloadJSON: payload}})
	for _, want := range []string{"## Gates", "verify_evidence", "task-ve-payload", "pass", "0.93", "validation bundle ok"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownAllGatesIncludingVerifyEvidence(t *testing.T) {
	taskID := "task-all-with-ve"
	payload := `{"gates":{"history":[
		{"gate":"enrich","status":"pass","confidence":0.9,"notes":["enrich ok"]},
		{"gate":"governance","status":"pass","confidence":0.85,"notes":["gov ok"]},
		{"gate":"human_review","status":"warn","confidence":1,"notes":["minor doc gap"]},
		{"gate":"verify_evidence","status":"pass","confidence":0.92,"notes":["evidence ok"]}
	]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: taskID, PayloadJSON: payload}})
	for _, want := range []string{"enrich", "governance", "human_review", "verify_evidence", taskID, "evidence ok"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownNoVerifyEvidenceWhenAbsent(t *testing.T) {
	payload := `{"gates":{"history":[{"gate":"governance","status":"pass","confidence":1}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: "task-no-ve", PayloadJSON: payload}})
	if contains(md, "| `verify_evidence` |") {
		t.Fatalf("verify_evidence gate row must not appear when absent from history:\n%s", md)
	}
}

func TestGatesMarkdownVerifyEvidencePayloadOverLogFallback(t *testing.T) {
	repo := t.TempDir()
	taskID := "task-ve-payload-win"
	payload := `{"gates":{"history":[{"gate":"verify_evidence","status":"pass","confidence":1,"notes":["from payload"]}]}}`
	logDir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := gates.NewLogDocument(taskID, "task", "verify_evidence", "feat", "reviewer", gates.Result{
		Status:     gates.VerdictFail,
		Confidence: 0.1,
		Notes:      []string{"from log only"},
	}, "2026-06-07T12:00:00Z")
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "verify_evidence.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	md := gatesMarkdown(repo, "", []sqlite.Task{{ID: taskID, PayloadJSON: payload}})
	if !contains(md, "pass") || !contains(md, "from payload") {
		t.Fatalf("expected payload to win:\n%s", md)
	}
	if contains(md, "from log only") || contains(md, "0.10") {
		t.Fatalf("log fallback must not override payload:\n%s", md)
	}
}

func TestGatesMarkdownVerifyEvidenceFromLogFallback(t *testing.T) {
	repo := t.TempDir()
	taskID := "task-ve-log"
	logDir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := gates.NewLogDocument(taskID, "task", "verify_evidence", "feat", "reviewer", gates.Result{
		Status:     gates.VerdictPass,
		Confidence: 0.91,
		Notes:      []string{"validation evidence ok"},
	}, "2026-06-07T12:00:00Z")
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "verify_evidence.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	md := gatesMarkdown(repo, "", []sqlite.Task{{ID: taskID, PayloadJSON: `{}`}})
	for _, want := range []string{"verify_evidence", taskID, "pass", "validation evidence ok"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownEnrichFromLogFallback(t *testing.T) {
	repo := t.TempDir()
	taskID := "task-enrich-log"
	logDir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := gates.NewLogDocument(taskID, "task", "enrich", "feat", "reviewer", gates.Result{
		Status:     gates.VerdictPass,
		Confidence: 0.88,
		Notes:      []string{"enrich gate ok"},
	}, "2026-06-07T12:00:00Z")
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "enrich.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	md := gatesMarkdown(repo, "", []sqlite.Task{{ID: taskID, PayloadJSON: `{}`}})
	for _, want := range []string{"enrich", taskID, "pass", "enrich gate ok"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownNoEnrichWhenAbsent(t *testing.T) {
	payload := `{"gates":{"history":[{"gate":"governance","status":"pass","confidence":1}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: "task-no-enrich", PayloadJSON: payload}})
	if contains(md, "| `enrich` |") {
		t.Fatalf("enrich gate row must not appear when absent from history:\n%s", md)
	}
}

func TestGatesMarkdownGovernanceHistoryFallback(t *testing.T) {
	payload := `{"title":"t1","governance":{"history":[{"status":"fail","confidence":0.3,"notes":["drift"]}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: "task-2", PayloadJSON: payload}})
	for _, want := range []string{"## Gates", "governance", "task-2", "fail", "drift"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownPrefersGatesOverGovernance(t *testing.T) {
	payload := `{"gates":{"history":[{"gate":"governance","status":"pass","confidence":1}]},"governance":{"history":[{"status":"fail","confidence":0}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: "task-3", PayloadJSON: payload}})
	if !contains(md, "pass") {
		t.Fatalf("expected gates.history to win:\n%s", md)
	}
}

func TestGatesMarkdownPlanFromLogFallback(t *testing.T) {
	repo := t.TempDir()
	runID := "run-plan-1"
	logDir := filepath.Join(repo, ".asagiri", "logs", runID, "gates")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := gates.NewLogDocument(runID, "run", "plan", "feat", "reviewer", gates.Result{
		Status:     gates.VerdictPass,
		Confidence: 0.95,
		Notes:      []string{"plan ok"},
	}, "2026-06-06T12:00:00Z")
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "plan.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	md := gatesMarkdown(repo, runID, nil)
	for _, want := range []string{"## Gates", "plan", runID, "pass", "plan ok"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownEmpty(t *testing.T) {
	if gatesMarkdown("", "", []sqlite.Task{{ID: "x", PayloadJSON: `{}`}}) != "" {
		t.Fatal("expected empty section")
	}
}

func TestGatesMarkdownPayloadOverLogFallback(t *testing.T) {
	repo := t.TempDir()
	taskID := "task-payload-win"
	payload := `{"gates":{"history":[{"gate":"governance","status":"pass","confidence":1,"notes":["from payload"]}]}}`
	logDir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := gates.NewLogDocument(taskID, "task", "governance", "feat", "reviewer", gates.Result{
		Status:     gates.VerdictFail,
		Confidence: 0.1,
		Notes:      []string{"from log only"},
	}, "2026-06-06T12:00:00Z")
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "governance.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	md := gatesMarkdown(repo, "run-x", []sqlite.Task{{ID: taskID, PayloadJSON: payload}})
	if !contains(md, "pass") || !contains(md, "from payload") {
		t.Fatalf("expected payload to win:\n%s", md)
	}
	if contains(md, "from log only") || contains(md, "0.10") {
		t.Fatalf("log fallback must not override payload:\n%s", md)
	}
}

func TestGatesMarkdownPlanAndGovernanceTogether(t *testing.T) {
	repo := t.TempDir()
	runID := "run-both"
	taskID := "task-gov"
	payload := `{"gates":{"history":[{"gate":"governance","status":"warn","confidence":0.8,"notes":["gov note"]}]}}`
	planDir := filepath.Join(repo, ".asagiri", "logs", runID, "gates")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planDoc := gates.NewLogDocument(runID, "run", "plan", "feat", "reviewer", gates.Result{
		Status:     gates.VerdictPass,
		Confidence: 0.99,
		Notes:      []string{"plan ok"},
	}, "2026-06-06T12:00:00Z")
	body, err := json.Marshal(planDoc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "plan.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	md := gatesMarkdown(repo, runID, []sqlite.Task{{ID: taskID, PayloadJSON: payload}})
	for _, want := range []string{"plan", runID, "pass", "plan ok", "governance", taskID, "warn", "gov note"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownGovernanceAndHumanReviewTogether(t *testing.T) {
	taskID := "task-both-gates"
	payload := `{"gates":{"history":[{"gate":"governance","status":"pass","confidence":0.9,"notes":["gov ok"]},{"gate":"human_review","status":"warn","confidence":1,"notes":["needs doc"]}]}}`
	md := gatesMarkdown("", "", []sqlite.Task{{ID: taskID, PayloadJSON: payload}})
	for _, want := range []string{"governance", "human_review", taskID, "pass", "gov ok", "warn", "needs doc"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGatesMarkdownHumanReviewFromLogFallback(t *testing.T) {
	repo := t.TempDir()
	taskID := "task-hr"
	logDir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := gates.NewLogDocument(taskID, "task", "human_review", "feat", "human", gates.Result{
		Status:     gates.VerdictPass,
		Confidence: 1,
		Notes:      []string{"approved"},
	}, "2026-06-06T12:00:00Z")
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "human_review.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	md := gatesMarkdown(repo, "", []sqlite.Task{{ID: taskID, PayloadJSON: `{}`}})
	for _, want := range []string{"human_review", taskID, "pass", "approved"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}
