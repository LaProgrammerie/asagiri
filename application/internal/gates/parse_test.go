package gates

import (
	"strings"
	"testing"
)

func TestParseResultYAMLPassWarnFail(t *testing.T) {
	cfg := ParseConfig{BlockKey: "governance", ParseErrorNote: "governance_parse_error"}
	cases := []struct {
		name   string
		stdout string
		want   Verdict
	}{
		{
			name: "pass",
			stdout: `governance:
  status: pass
  confidence: 0.95
`,
			want: VerdictPass,
		},
		{
			name: "warn",
			stdout: `governance:
  status: warn
  confidence: 0.7
  findings:
    - code: other
      severity: warn
      message: minor drift
`,
			want: VerdictWarn,
		},
		{
			name: "fail",
			stdout: `governance:
  status: fail
  confidence: 0.4
  findings:
    - code: spec_drift
      severity: fail
      message: spec mismatch
`,
			want: VerdictFail,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := ParseResult(tc.stdout, cfg)
			got := ClassifyResult(r, nil).Status
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestParseResultJSONPassFail(t *testing.T) {
	cfg := ParseConfig{BlockKey: "governance", ParseErrorNote: "governance_parse_error"}

	passStdout := `{"governance":{"status":"pass","confidence":0.88,"notes":["json ok"]}}`
	r := ParseResult(passStdout, cfg)
	got := ClassifyResult(r, nil)
	if got.Status != VerdictPass {
		t.Fatalf("json pass: got %q", got.Status)
	}
	if got.Confidence != 0.88 {
		t.Fatalf("confidence: got %v", got.Confidence)
	}

	failStdout := `{"governance":{"status":"fail","findings":[{"code":"spec_drift","severity":"fail","message":"json drift"}]}}`
	r = ParseResult(failStdout, cfg)
	got = ClassifyResult(r, []string{"spec_drift"})
	if got.Status != VerdictFail {
		t.Fatalf("json fail: got %q", got.Status)
	}
}

func TestParseResultInvalidIsFailWithParseError(t *testing.T) {
	cfg := ParseConfig{
		BlockKey:          "governance",
		MissingBlockError: "governance block missing from agent output",
		ParseErrorNote:    "governance_parse_error",
	}
	r := ParseResult("not yaml at all", cfg)
	got := ClassifyResult(r, nil)
	if got.Status != VerdictFail {
		t.Fatalf("malformed: got %q want fail", got.Status)
	}
	if r.ParseError == "" {
		t.Fatal("expected parse_error")
	}
	if len(r.Notes) == 0 || r.Notes[0] != "governance_parse_error" {
		t.Fatalf("notes: got %v", r.Notes)
	}
}

func TestParseResultMissingBlockIsFail(t *testing.T) {
	cfg := ParseConfig{
		BlockKey:          "governance",
		MissingBlockError: "governance block missing from agent output",
		ParseErrorNote:    "governance_parse_error",
	}
	r := ParseResult("", cfg)
	got := ClassifyResult(r, nil)
	if got.Status != VerdictFail {
		t.Fatalf("missing block: got %q want fail", got.Status)
	}
}

func TestClassifyFindingSeverityFailOverridesPass(t *testing.T) {
	r := Result{
		Status: VerdictPass,
		Findings: []Finding{
			{Code: "spec_drift", Severity: "fail", Message: "drift"},
		},
	}
	got := ClassifyResult(r, []string{"spec_drift"})
	if got.Status != VerdictFail {
		t.Fatalf("finding fail: got %q want fail", got.Status)
	}
}

func TestClassifyFailOnFilter(t *testing.T) {
	r := Result{
		Status: VerdictPass,
		Findings: []Finding{
			{Code: "architecture_violation", Severity: "fail", Message: "violation"},
		},
	}
	got := ClassifyResult(r, []string{"spec_drift"})
	if got.Status != VerdictPass {
		t.Fatalf("filtered fail code: got %q want pass", got.Status)
	}
}

func TestParseStatusAliases(t *testing.T) {
	cfg := ParseConfig{BlockKey: "governance"}
	cases := []struct {
		status string
		want   Verdict
	}{
		{"PASS", VerdictPass},
		{"WARN", VerdictWarn},
		{"FAIL", VerdictFail},
		{"passed", VerdictPass},
		{"warning", VerdictWarn},
		{"failed", VerdictFail},
	}
	for _, tc := range cases {
		stdout := "governance:\n  status: " + tc.status + "\n  confidence: 0.5\n"
		r := ParseResult(stdout, cfg)
		got := ClassifyResult(r, nil).Status
		if got != tc.want {
			t.Fatalf("status %q: got %q want %q", tc.status, got, tc.want)
		}
	}
}

func TestFormatFailure(t *testing.T) {
	msg := FormatFailure(Result{
		Findings: []Finding{
			{Code: "spec_drift", Severity: "fail", Message: "drift", Actions: []string{"fix spec"}},
		},
	})
	if !strings.Contains(msg, "spec_drift") || !strings.Contains(msg, "fix spec") {
		t.Fatalf("format: %q", msg)
	}
}
