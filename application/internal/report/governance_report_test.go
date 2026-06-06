package report

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

func TestGovernanceMarkdown(t *testing.T) {
	payload := `{"title":"t1","governance":{"history":[{"status":"warn","confidence":0.72,"notes":["API drift"]}]}}`
	md := governanceMarkdown([]sqlite.Task{{ID: "task-1", PayloadJSON: payload}})
	for _, want := range []string{"## Governance", "task-1", "warn", "0.72", "API drift"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func TestGovernanceMarkdownEmpty(t *testing.T) {
	if governanceMarkdown([]sqlite.Task{{ID: "x", PayloadJSON: `{}`}}) != "" {
		t.Fatal("expected empty section")
	}
}
