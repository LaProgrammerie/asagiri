package workflow

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const governanceMaxExcerpt = 12000

var governanceParseConfig = gates.ParseConfig{
	BlockKey:          "governance",
	MissingBlockError: "governance block missing from agent output",
	ParseErrorNote:    "governance_parse_error",
}

// governanceLogDocument is the legacy JSON shape; new logs use gateLogDocument under .asagiri/logs/<task>/gates/governance.json.
type governanceLogDocument struct {
	TaskID     string                      `json:"task_id"`
	Feature    string                      `json:"feature"`
	At         string                      `json:"at"`
	Status     string                      `json:"status"`
	Confidence float64                     `json:"confidence"`
	Notes      []string                    `json:"notes,omitempty"`
	Findings   []asagiri.GovernanceFinding `json:"findings,omitempty"`
	Evidence   []gates.EvidenceRef         `json:"evidence,omitempty"`
	DryRun     bool                        `json:"dry_run,omitempty"`
	ParseError string                      `json:"parse_error,omitempty"`
	Agent      string                      `json:"agent,omitempty"`
}

func parseGovernanceVerdict(stdout string) gates.Result {
	return gates.ParseResult(stdout, governanceParseConfig)
}

func classifyGovernanceVerdict(r gates.Result, failOn []string) string {
	return string(gates.ClassifyResult(r, failOn).Status)
}

func truncateGovernanceText(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "\n… [truncated]"
}
