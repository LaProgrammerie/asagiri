package workflow

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
)

const governanceMaxExcerpt = 12000

var governanceParseConfig = gates.ParseConfig{
	BlockKey:          "governance",
	MissingBlockError: "governance block missing from agent output",
	ParseErrorNote:    "governance_parse_error",
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
