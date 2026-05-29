package coordination

import "context"

// ConflictCategory classifies detected coordination conflicts (spec-my-D §15).
type ConflictCategory string

const (
	ConflictConcurrentEdit ConflictCategory = "concurrent_edit"
	ConflictFileOverlap    ConflictCategory = "file_overlap"
	ConflictContractDrift  ConflictCategory = "contract_drift"
	ConflictFlowDrift      ConflictCategory = "flow_drift"
	ConflictTrustDowngrade ConflictCategory = "trust_downgrade"
)

// Conflict describes one detected issue.
type Conflict struct {
	Category ConflictCategory `json:"category"`
	Message  string           `json:"message"`
	NodeIDs  []string         `json:"node_ids,omitempty"`
}

// ConflictDetector finds cross-agent conflicts (Lot 4 stub).
type ConflictDetector interface {
	Detect(ctx context.Context, graph ExecutionGraph) ([]Conflict, error)
}

// StubConflictDetector is a Lot-1 placeholder.
type StubConflictDetector struct{}

func (StubConflictDetector) Detect(_ context.Context, _ ExecutionGraph) ([]Conflict, error) {
	return nil, ErrNotImplemented
}
