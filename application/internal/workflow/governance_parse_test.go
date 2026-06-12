package workflow

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
)

func TestParseGovernanceVerdictPass(t *testing.T) {
	stdout := `governance:
  status: pass
  confidence: 0.95
  notes:
    - ok
`
	v := parseGovernanceVerdict(stdout)
	got := classifyGovernanceVerdict(v, nil)
	if got != "pass" {
		t.Fatalf("status: got %q want pass", got)
	}
}

func TestParseGovernanceVerdictWarn(t *testing.T) {
	stdout := `governance:
  status: warn
  confidence: 0.7
  findings:
    - code: other
      severity: warn
      message: minor drift
`
	v := parseGovernanceVerdict(stdout)
	got := classifyGovernanceVerdict(v, nil)
	if got != "warn" {
		t.Fatalf("status: got %q want warn", got)
	}
}

func TestParseGovernanceVerdictFail(t *testing.T) {
	stdout := `governance:
  status: fail
  confidence: 0.4
  findings:
    - code: spec_drift
      severity: fail
      message: spec mismatch
`
	v := parseGovernanceVerdict(stdout)
	got := classifyGovernanceVerdict(v, []string{"spec_drift"})
	if got != "fail" {
		t.Fatalf("status: got %q want fail", got)
	}
}

func TestParseGovernanceMalformedIsFail(t *testing.T) {
	v := parseGovernanceVerdict("not yaml at all")
	got := classifyGovernanceVerdict(v, nil)
	if got != "fail" {
		t.Fatalf("malformed: got %q want fail", got)
	}
	if v.ParseError == "" {
		t.Fatal("expected parse_error")
	}
	if len(v.Notes) == 0 || v.Notes[0] != "governance_parse_error" {
		t.Fatalf("notes: got %v want governance_parse_error", v.Notes)
	}
}

func TestParseGovernanceMissingBlockIsFail(t *testing.T) {
	v := parseGovernanceVerdict("")
	got := classifyGovernanceVerdict(v, nil)
	if got != "fail" {
		t.Fatalf("missing block: got %q want fail", got)
	}
}

func TestClassifyFindingSeverityFailOverridesPass(t *testing.T) {
	v := gates.Result{
		Status: gates.VerdictPass,
		Findings: []gates.Finding{
			{Code: "spec_drift", Severity: "fail", Message: "drift"},
		},
	}
	got := classifyGovernanceVerdict(v, []string{"spec_drift"})
	if got != "fail" {
		t.Fatalf("finding fail: got %q want fail", got)
	}
}

func TestClassifyFailOnFilter(t *testing.T) {
	v := gates.Result{
		Status: gates.VerdictPass,
		Findings: []gates.Finding{
			{Code: "architecture_violation", Severity: "fail", Message: "violation"},
		},
	}
	got := classifyGovernanceVerdict(v, []string{"spec_drift"})
	if got != "pass" {
		t.Fatalf("filtered fail code: got %q want pass", got)
	}
}

func TestParseGovernanceStatusAliases(t *testing.T) {
	cases := []struct {
		name   string
		status string
		want   string
	}{
		{"PASS uppercase", "PASS", "pass"},
		{"WARN uppercase", "WARN", "warn"},
		{"FAIL uppercase", "FAIL", "fail"},
		{"passed alias", "passed", "pass"},
		{"warning alias", "warning", "warn"},
		{"failed alias", "failed", "fail"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout := "governance:\n  status: " + tc.status + "\n  confidence: 0.5\n"
			v := parseGovernanceVerdict(stdout)
			got := classifyGovernanceVerdict(v, nil)
			if got != tc.want {
				t.Fatalf("status %q: got %q want %q (parsed status %q)", tc.status, got, tc.want, v.Status)
			}
		})
	}
}

func TestParseGovernanceVerdictJSON(t *testing.T) {
	stdout := `{"governance":{"status":"pass","confidence":0.88,"notes":["json ok"]}}`
	v := parseGovernanceVerdict(stdout)
	got := classifyGovernanceVerdict(v, nil)
	if got != "pass" {
		t.Fatalf("json pass: got %q want pass", got)
	}
	if v.Confidence != 0.88 {
		t.Fatalf("confidence: got %v want 0.88", v.Confidence)
	}
	if len(v.Notes) != 1 || v.Notes[0] != "json ok" {
		t.Fatalf("notes: got %v", v.Notes)
	}
}

func TestParseGovernanceVerdictJSONFailFinding(t *testing.T) {
	stdout := `{
  "governance": {
    "status": "fail",
    "confidence": 0.3,
    "findings": [
      {"code": "spec_drift", "severity": "fail", "message": "json drift"}
    ]
  }
}`
	v := parseGovernanceVerdict(stdout)
	got := classifyGovernanceVerdict(v, []string{"spec_drift"})
	if got != "fail" {
		t.Fatalf("json fail: got %q want fail", got)
	}
}
