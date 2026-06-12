package workflow

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
)

func TestGovernanceEntryRecordParity(t *testing.T) {
	r := gates.Result{
		Status:     gates.VerdictWarn,
		Confidence: 0.5,
		Notes:      []string{"note"},
		Findings: []gates.Finding{
			{Code: "spec_drift", Severity: "warn", Message: "drift", Actions: []string{"fix"}},
		},
		DryRun:     true,
		ParseError: "parse err",
	}
	at := "2026-06-06T12:00:00Z"
	retry := 2

	entry := gateHistoryEntryFromResult("governance", r, at, retry)
	rec := governanceRecordFromEntry(entry)

	if entry.Gate != "governance" {
		t.Fatalf("gate: %q", entry.Gate)
	}
	if rec.At != entry.At || rec.Status != entry.Status || rec.Confidence != entry.Confidence ||
		rec.Retry != entry.Retry || rec.DryRun != entry.DryRun || rec.ParseError != entry.ParseError {
		t.Fatalf("metadata mismatch entry=%+v record=%+v", entry, rec)
	}
	if len(rec.Notes) != len(entry.Notes) || rec.Notes[0] != entry.Notes[0] {
		t.Fatalf("notes mismatch entry=%+v record=%+v", entry.Notes, rec.Notes)
	}
	if len(rec.Findings) != 1 || rec.Findings[0].Code != entry.Findings[0].Code ||
		rec.Findings[0].Severity != entry.Findings[0].Severity ||
		rec.Findings[0].Message != entry.Findings[0].Message ||
		len(rec.Findings[0].Actions) != len(entry.Findings[0].Actions) {
		t.Fatalf("findings mismatch entry=%+v record=%+v", entry.Findings, rec.Findings)
	}
}

func TestGovernanceRecordFromResult(t *testing.T) {
	r := gates.Result{
		GateID:     "governance",
		Status:     gates.VerdictWarn,
		Confidence: 0.5,
		Notes:      []string{"note"},
		Findings: []gates.Finding{
			{Code: "spec_drift", Severity: "warn", Message: "drift"},
		},
		DryRun: true,
	}
	rec := governanceRecordFromResult(r, "2026-06-06T12:00:00Z", 2)
	if rec.Status != "warn" || rec.Retry != 2 || !rec.DryRun {
		t.Fatalf("record: %+v", rec)
	}
	if len(rec.Findings) != 1 || rec.Findings[0].Code != "spec_drift" {
		t.Fatalf("findings: %+v", rec.Findings)
	}
}
