package onboarding

import "github.com/LaProgrammerie/asagiri/application/internal/config"

const (
	dirRel       = ".asagiri/onboarding"
	stateRel     = dirRel + "/state.json"
	reportRel    = dirRel + "/report.json"
	backupsRel   = dirRel + "/backups"
	scoreFail    = 20
	scoreWarn    = 5
	initialScore = 100
)

// CheckStatus is one readiness/doctor check outcome.
type CheckStatus string

const (
	StatusOK   CheckStatus = "ok"
	StatusWarn CheckStatus = "warn"
	StatusFail CheckStatus = "fail"
)

// Check is one readiness or doctor item.
type Check struct {
	ID      string      `json:"id"`
	Status  CheckStatus `json:"status"`
	Message string      `json:"message,omitempty"`
	FixCLI  string      `json:"fix_cli,omitempty"`
}

// Action suggests a next step for the operator.
type Action struct {
	Title string `json:"title"`
	CLI   string `json:"cli"`
}

// Report is the readiness score and checklist.
type Report struct {
	Ready       bool     `json:"ready"`
	Score       int      `json:"score"`
	Checks      []Check  `json:"checks"`
	NextActions []Action `json:"next_actions"`
}

// Options controls onboard/ready/doctor behaviour.
type Options struct {
	Yes             bool
	NonInteractive  bool
	Stack           string
	Resume          bool
	Step            string
	Plain           bool
	JSON            bool
	CI              bool
	DryRun          bool
	ForceDocs       bool
	Strict          bool
	CheckOnly       bool
	UI              bool
	Autofix         bool
	ProjectName     string
	DefaultBranch   string
	Tagline         string
	DefaultAgent    string
	DefaultReviewer string
	FeatureSlug     string
	ProductOneLiner string
	ProductUsers    string
}

// Answers collects wizard inputs persisted in state.
type Answers struct {
	ProjectName     string `json:"project_name,omitempty"`
	DefaultBranch   string `json:"default_branch,omitempty"`
	Tagline         string `json:"tagline,omitempty"`
	Stack           string `json:"stack,omitempty"`
	DefaultAgent    string `json:"default_agent,omitempty"`
	DefaultReviewer string `json:"default_reviewer,omitempty"`
	FeatureSlug     string `json:"feature_slug,omitempty"`
	ProductOneLiner string `json:"product_one_liner,omitempty"`
	ProductUsers    string `json:"product_users,omitempty"`
}

// PlannedChange describes one file write during dry-run.
type PlannedChange struct {
	Path    string `json:"path"`
	Action  string `json:"action"`
	Summary string `json:"summary,omitempty"`
}

// Result summarizes an onboard run.
type Result struct {
	Report         Report          `json:"report,omitempty"`
	PlannedChanges []PlannedChange `json:"planned_changes,omitempty"`
	SkippedFields  []string        `json:"skipped_fields,omitempty"`
	ConfigPath     string          `json:"config_path,omitempty"`
	BackupPath     string          `json:"backup_path,omitempty"`
}

// ConfigPatch holds fields to merge into config.yaml.
type ConfigPatch struct {
	ProjectName     string
	DefaultBranch   string
	BranchPrefix    string
	DefaultAgent    string
	DefaultReviewer string
	Validation      []config.ValidationCommand
}
