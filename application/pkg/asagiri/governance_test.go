package asagiri

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTaskGovernanceJSONRoundTrip(t *testing.T) {
	task := Task{
		ID:      "task-gov",
		Title:   "Governance round-trip",
		Feature: "feat",
		Status:  StatusImplemented,
		Governance: &TaskGovernance{
			Retries: 1,
			History: []GovernanceRecord{
				{
					At:         "2026-06-06T12:00:00Z",
					Status:     "fail",
					Confidence: 0.2,
					Retry:      0,
					Findings: []GovernanceFinding{
						{Code: "spec_drift", Severity: "fail", Message: "drift"},
					},
				},
				{
					At:         "2026-06-06T12:05:00Z",
					Status:     "pass",
					Confidence: 1,
					Retry:      1,
					Notes:      []string{"ok"},
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
	if back.Governance == nil {
		t.Fatal("governance missing after round-trip")
	}
	if back.Governance.Retries != 1 {
		t.Fatalf("retries: got %d want 1", back.Governance.Retries)
	}
	if len(back.Governance.History) != 2 {
		t.Fatalf("history len: got %d want 2", len(back.Governance.History))
	}
	if back.Governance.History[0].Retry != 0 || back.Governance.History[1].Status != "pass" {
		t.Fatalf("history: %+v", back.Governance.History)
	}
}

func TestTaskGovernanceYAMLRoundTrip(t *testing.T) {
	task := Task{
		ID:      "task-gov-yaml",
		Title:   "YAML",
		Feature: "feat",
		Governance: &TaskGovernance{
			Retries: 2,
			History: []GovernanceRecord{
				{At: "2026-06-06T12:00:00Z", Status: "fail", Retry: 0},
				{At: "2026-06-06T12:01:00Z", Status: "fail", Retry: 1},
				{At: "2026-06-06T12:02:00Z", Status: "fail", Retry: 2},
			},
		},
	}

	body, err := yaml.Marshal(task)
	if err != nil {
		t.Fatalf("marshal yaml: %v", err)
	}
	var back Task
	if err := yaml.Unmarshal(body, &back); err != nil {
		t.Fatalf("unmarshal yaml: %v", err)
	}
	if back.Governance == nil || back.Governance.Retries != 2 || len(back.Governance.History) != 3 {
		t.Fatalf("yaml round-trip: %+v", back.Governance)
	}
}
