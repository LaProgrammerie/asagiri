package investigation

import "time"

// Depth controls investigation thoroughness (spec-my-A §25.4).
type Depth string

const (
	DepthQuick     Depth = "quick"
	DepthStandard  Depth = "standard"
	DepthDeep      Depth = "deep"
	DepthCI        Depth = "ci"
)

// Request is the structured investigation input.
type Request struct {
	Symptom          string
	Feature          string
	Flow             string
	TaskID           string
	RunID            string
	FromFailedTests  bool
	Depth            Depth
	MaxFiles         int
	MaxDuration      time.Duration
	NoCloud          bool
	EstimateOnly     bool
	Output           string // markdown | json
	RepoRoot         string
}
