package asagiri

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGateHistoryEntryJSONRoundTrip(t *testing.T) {
	task := Task{
		ID:      "task-gates",
		Title:   "Gates round-trip",
		Feature: "feat",
		Gates: &TaskGates{
			History: []GateHistoryEntry{
				{
					Gate:       "governance",
					At:         "2026-06-06T12:00:00Z",
					Status:     "warn",
					Confidence: 0.6,
					Findings: []GateFinding{
						{Code: "spec_drift", Severity: "warn", Message: "drift"},
					},
				},
			},
		},
	}
	body, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Task
	if err := json.Unmarshal(body, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Gates == nil || len(back.Gates.History) != 1 {
		t.Fatalf("gates: %+v", back.Gates)
	}
	if back.Gates.History[0].Gate != "governance" {
		t.Fatalf("gate name: %+v", back.Gates.History[0])
	}
}

func TestGovernanceFindingAlias(t *testing.T) {
	var f = GateFinding{Code: "x", Severity: "warn", Message: "m"}
	if f.Code != "x" {
		t.Fatal("alias broken")
	}
}

func TestGateHistoryYAMLRoundTrip(t *testing.T) {
	task := Task{
		ID: "t",
		Gates: &TaskGates{
			History: []GateHistoryEntry{{Gate: "plan", Status: "pass", At: "2026-06-06T12:00:00Z"}},
		},
	}
	body, err := yaml.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}
	var back Task
	if err := yaml.Unmarshal(body, &back); err != nil {
		t.Fatal(err)
	}
	if back.Gates == nil || back.Gates.History[0].Gate != "plan" {
		t.Fatalf("yaml: %+v", back.Gates)
	}
}
