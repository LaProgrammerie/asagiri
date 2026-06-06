package workflow

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

const governanceMaxExcerpt = 12000

// governanceVerdict is the parsed agent output before classification.
type governanceVerdict struct {
	Status     string                    `yaml:"status"`
	Confidence float64                   `yaml:"confidence"`
	Notes      []string                  `yaml:"notes"`
	Findings   []asagiri.GovernanceFinding `yaml:"findings"`
	DryRun     bool                      `yaml:"dry_run,omitempty"`
	ParseError string                    `yaml:"parse_error,omitempty"`
}

// governanceLogDocument is written to .asagiri/logs/<task>/governance.json.
type governanceLogDocument struct {
	TaskID     string                      `json:"task_id"`
	Feature    string                      `json:"feature"`
	At         string                      `json:"at"`
	Status     string                      `json:"status"`
	Confidence float64                     `json:"confidence"`
	Notes      []string                    `json:"notes,omitempty"`
	Findings   []asagiri.GovernanceFinding `json:"findings,omitempty"`
	DryRun     bool                        `json:"dry_run,omitempty"`
	ParseError string                      `json:"parse_error,omitempty"`
	Agent      string                      `json:"agent,omitempty"`
}

func parseGovernanceVerdict(stdout string) governanceVerdict {
	raw := extractGovernancePayload(stdout)
	if raw == "" {
		return governanceVerdict{
			Status:     "fail",
			ParseError: "governance block missing from agent output",
			Notes:      []string{"governance_parse_error"},
		}
	}

	var wrapper struct {
		Governance governanceVerdict `yaml:"governance" json:"governance"`
	}
	if err := yaml.Unmarshal([]byte(raw), &wrapper); err != nil {
		if jsonErr := json.Unmarshal([]byte(raw), &wrapper); jsonErr != nil {
			return governanceVerdict{
				Status:     "fail",
				ParseError: fmt.Sprintf("governance parse error: %v", err),
				Notes:      []string{"governance_parse_error"},
			}
		}
	}
	v := wrapper.Governance
	v.Status = normalizeGovernanceStatus(v.Status)
	v.Confidence = clampConfidence(v.Confidence)
	for i := range v.Findings {
		v.Findings[i].Severity = normalizeGovernanceSeverity(v.Findings[i].Severity)
		v.Findings[i].Code = strings.TrimSpace(v.Findings[i].Code)
	}
	return v
}

func classifyGovernanceVerdict(v governanceVerdict, failOn []string) string {
	if strings.TrimSpace(v.ParseError) != "" {
		return "fail"
	}
	blocking := failOnSet(failOn)
	hasWarn := false
	for _, f := range v.Findings {
		switch f.Severity {
		case "fail":
			code := strings.ToLower(strings.TrimSpace(f.Code))
			if len(blocking) == 0 || blocking[code] || code == "" {
				return "fail"
			}
		case "warn":
			hasWarn = true
		}
	}
	if hasWarn {
		return "warn"
	}
	if v.Status == "fail" || v.Status == "warn" || v.Status == "pass" {
		return v.Status
	}
	return "pass"
}

func failOnSet(codes []string) map[string]bool {
	if len(codes) == 0 {
		return nil
	}
	out := make(map[string]bool, len(codes))
	for _, c := range codes {
		c = strings.ToLower(strings.TrimSpace(c))
		if c != "" {
			out[c] = true
		}
	}
	return out
}

func normalizeGovernanceStatus(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pass", "passed", "ok":
		return "pass"
	case "warn", "warning":
		return "warn"
	case "fail", "failed", "error":
		return "fail"
	default:
		return strings.ToLower(strings.TrimSpace(s))
	}
}

func normalizeGovernanceSeverity(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "fail", "failed", "error", "block", "blocking":
		return "fail"
	case "warn", "warning", "advisory":
		return "warn"
	default:
		return strings.ToLower(strings.TrimSpace(s))
	}
}

func clampConfidence(c float64) float64 {
	if c < 0 {
		return 0
	}
	if c > 1 {
		return 1
	}
	return c
}

func extractGovernancePayload(stdout string) string {
	s := strings.TrimSpace(stdout)
	if s == "" {
		return ""
	}
	if fenced := extractYAMLFence(s); fenced != "" {
		s = fenced
	}
	idx := strings.Index(strings.ToLower(s), "governance:")
	if idx >= 0 {
		return strings.TrimSpace(s[idx:])
	}
	return s
}

func extractYAMLFence(s string) string {
	lower := strings.ToLower(s)
	start := strings.Index(lower, "```yaml")
	if start < 0 {
		start = strings.Index(lower, "```yml")
	}
	if start < 0 {
		return ""
	}
	rest := s[start:]
	if i := strings.Index(rest, "\n"); i >= 0 {
		rest = rest[i+1:]
	}
	end := strings.Index(rest, "```")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func truncateGovernanceText(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "\n… [truncated]"
}

func governanceSummary(v governanceVerdict) string {
	if v.ParseError != "" {
		return v.ParseError
	}
	if len(v.Notes) > 0 {
		return v.Notes[0]
	}
	if len(v.Findings) > 0 {
		return v.Findings[0].Message
	}
	return "governance gate failed"
}

func formatGovernanceFailure(v governanceVerdict) string {
	var parts []string
	if msg := governanceSummary(v); msg != "" {
		parts = append(parts, msg)
	}
	for _, f := range v.Findings {
		line := fmt.Sprintf("[%s/%s] %s", f.Code, f.Severity, f.Message)
		if len(f.Actions) > 0 {
			line += " — actions: " + strings.Join(f.Actions, "; ")
		}
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return "governance gate failed"
	}
	return strings.Join(parts, " | ")
}
