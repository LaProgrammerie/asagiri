package replay

import (
	"time"
)

const (
	RelDir          = ".asagiri/replays"
	ManifestName    = "replay.yaml"
	SnapshotsRelDir = "snapshots"
)

// Mode identifies a replay execution mode (spec §11).
type Mode string

const (
	ModeFull       Mode = "full"
	ModeSimulation Mode = "simulation"
	ModeOffline    Mode = "offline"
	ModeAudit      Mode = "audit"
	ModeCompare    Mode = "compare"
)

// Manifest is the replay.yaml schema (spec §7).
type Manifest struct {
	ID        string            `yaml:"id"`
	CreatedAt time.Time         `yaml:"created_at"`
	Source    SourceRef         `yaml:"source"`
	Repo      RepoRef           `yaml:"repo"`
	Runtime   RuntimeRef        `yaml:"runtime"`
	Agents    map[string]string `yaml:"agents,omitempty"`
	Policies  map[string]any    `yaml:"policies,omitempty"`
	Artifacts []string          `yaml:"artifacts,omitempty"`
}

// SourceRef links the replay to its origin workflow.
type SourceRef struct {
	Run           string `yaml:"run,omitempty"`
	Graph         string `yaml:"graph,omitempty"`
	Investigation string `yaml:"investigation,omitempty"`
}

// RepoRef captures git context (spec §10).
type RepoRef struct {
	Commit string `yaml:"commit,omitempty"`
	Branch string `yaml:"branch,omitempty"`
}

// RuntimeRef captures runtime version and mode.
type RuntimeRef struct {
	AsagiriVersion string `yaml:"asagiri_version,omitempty"`
	RuntimeMode    string `yaml:"runtime_mode,omitempty"`
}

// ReplayPackage is a persisted replay bundle on disk.
type ReplayPackage struct {
	ID       string
	Path     string
	Manifest Manifest
}

// ReplayCreateRequest configures package capture (spec §6.1).
type ReplayCreateRequest struct {
	RepoRoot          string
	Config            CapturePolicies
	FromRun           string
	FromGraph         string
	FromInvestigation string
	IncludeRuntime    bool
	IncludePrompts    bool
	IncludeEvents     bool
}

// ReplayRunRequest configures replay execution (spec §6.2).
type ReplayRunRequest struct {
	RepoRoot   string
	ReplayID   string
	Mode       Mode
	DryRun     bool
	Compare    bool
	Strict     bool
	Offline    bool
	Simulation bool
}

// ReplayResult summarizes a replay session.
type ReplayResult struct {
	ReplayID    string            `json:"replay_id"`
	Mode        Mode              `json:"mode"`
	Offline     bool              `json:"offline"`
	Artifacts   map[string]bool   `json:"artifacts"`
	Warnings    []string          `json:"warnings,omitempty"`
	Divergences []Divergence      `json:"divergences,omitempty"`
	Comparison  *ReplayComparison `json:"comparison,omitempty"`
}

// ReplayComparison holds diff between two replay packages (spec §14).
type ReplayComparison struct {
	ReplayA       string             `json:"replay_a"`
	ReplayB       string             `json:"replay_b"`
	CostDelta     float64            `json:"cost_delta,omitempty"`
	DurationDelta string             `json:"duration_delta,omitempty"`
	TrustDiff     map[string]float64 `json:"trust_diff,omitempty"`
	Differences   []string           `json:"differences,omitempty"`
	Divergences   []Divergence       `json:"divergences,omitempty"`
	Warnings      []string           `json:"warnings,omitempty"`
}
