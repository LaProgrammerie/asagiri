package gates

// Finding is a single structured issue reported by a gate validator.
type Finding struct {
	Code     string   `yaml:"code" json:"code"`
	Severity string   `yaml:"severity" json:"severity"`
	Message  string   `yaml:"message" json:"message"`
	Actions  []string `yaml:"actions,omitempty" json:"actions,omitempty"`
}

// EvidenceRef points to material consulted or produced during gate evaluation.
type EvidenceRef struct {
	Kind string `yaml:"kind" json:"kind"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	Note string `yaml:"note,omitempty" json:"note,omitempty"`
}

// Result is the structured output of a gate evaluation.
type Result struct {
	GateID     string        `yaml:"gate_id,omitempty" json:"gate_id,omitempty"`
	GateType   string        `yaml:"gate_type,omitempty" json:"gate_type,omitempty"`
	Scope      string        `yaml:"scope,omitempty" json:"scope,omitempty"`
	Status     Verdict       `yaml:"status" json:"status"`
	Confidence float64       `yaml:"confidence" json:"confidence"`
	Notes      []string      `yaml:"notes,omitempty" json:"notes,omitempty"`
	Findings   []Finding     `yaml:"findings,omitempty" json:"findings,omitempty"`
	Evidence   []EvidenceRef `yaml:"evidence,omitempty" json:"evidence,omitempty"`
	Retry      int           `yaml:"retry,omitempty" json:"retry,omitempty"`
	DryRun     bool          `yaml:"dry_run,omitempty" json:"dry_run,omitempty"`
	ParseError string        `yaml:"parse_error,omitempty" json:"parse_error,omitempty"`
}
