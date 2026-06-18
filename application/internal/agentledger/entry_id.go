package agentledger

import (
	"strings"
)

// LocalEntryID returns a stable idempotency key for one ledger line.
//
// Strategy: SHA-256 hex over a canonical tuple
// (run_id, agent_id, phase, started_at, prompt_hash, output_hash).
// The same ledger entry always yields the same id on every cloud push,
// without persisting an extra field in the JSONL file.
func LocalEntryID(e Entry) string {
	parts := []string{
		strings.TrimSpace(e.RunID),
		strings.TrimSpace(e.AgentID),
		strings.TrimSpace(e.Phase),
		strings.TrimSpace(e.StartedAt),
		strings.TrimSpace(e.PromptHash),
		strings.TrimSpace(e.OutputHash),
	}
	return HashText(strings.Join(parts, "\x1e"))
}
