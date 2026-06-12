package gates

// LogDocument is the JSON artefact written under .asagiri/logs/<scope-id>/gates/<gate>.json.
type LogDocument struct {
	ScopeID    string        `json:"scope_id"`
	ScopeKind  string        `json:"scope_kind"`
	GateName   string        `json:"gate_name"`
	RunID      string        `json:"run_id,omitempty"`
	TaskID     string        `json:"task_id,omitempty"`
	Feature    string        `json:"feature,omitempty"`
	At         string        `json:"at"`
	Status     string        `json:"status"`
	Confidence float64       `json:"confidence"`
	Notes      []string      `json:"notes,omitempty"`
	Findings   []Finding     `json:"findings,omitempty"`
	Evidence   []EvidenceRef `json:"evidence,omitempty"`
	DryRun     bool          `json:"dry_run,omitempty"`
	ParseError string        `json:"parse_error,omitempty"`
	Agent      string        `json:"agent,omitempty"`
}

// NewLogDocument builds a log document from a gate evaluation result.
func NewLogDocument(scopeID, scopeKind, gateName, feature, agentName string, r Result, at string) LogDocument {
	doc := LogDocument{
		ScopeID:    scopeID,
		ScopeKind:  scopeKind,
		GateName:   gateName,
		Feature:    feature,
		At:         at,
		Status:     string(r.Status),
		Confidence: r.Confidence,
		Notes:      r.Notes,
		Findings:   r.Findings,
		Evidence:   r.Evidence,
		DryRun:     r.DryRun,
		ParseError: r.ParseError,
		Agent:      agentName,
	}
	switch scopeKind {
	case "run":
		doc.RunID = scopeID
	case "task":
		doc.TaskID = scopeID
	}
	return doc
}
