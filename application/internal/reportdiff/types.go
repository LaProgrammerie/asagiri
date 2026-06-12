package reportdiff

const ReportVersion = "report-diff-v1"

type ScoreDelta struct {
	Before float64 `json:"before"`
	After  float64 `json:"after"`
	Delta  float64 `json:"delta"`
}

type VerdictDelta struct {
	Before  string `json:"before"`
	After   string `json:"after"`
	Changed bool   `json:"changed"`
}

type DimensionDelta struct {
	ID      string  `json:"id"`
	Label   string  `json:"label,omitempty"`
	Before  float64 `json:"before"`
	After   float64 `json:"after"`
	Delta   float64 `json:"delta"`
	Changed bool    `json:"changed"`
}

type NextActionDelta struct {
	BeforeCommand string `json:"before_command"`
	AfterCommand  string `json:"after_command"`
	Changed       bool   `json:"changed"`
}

type ReportPaths struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

type TrustTaskDiff struct {
	ReportVersion string           `json:"report_version"`
	Scope         string           `json:"scope"`
	ScopeID       string           `json:"scope_id"`
	Paths         ReportPaths      `json:"paths"`
	Score         ScoreDelta       `json:"score"`
	Verdict       VerdictDelta     `json:"verdict"`
	Dimensions    []DimensionDelta `json:"dimensions"`
	NextAction    NextActionDelta  `json:"next_action"`
}

type TrustFeatureDiff struct {
	ReportVersion  string          `json:"report_version"`
	Scope          string          `json:"scope"`
	ScopeID        string          `json:"scope_id"`
	Paths          ReportPaths     `json:"paths"`
	Score          ScoreDelta      `json:"score"`
	Verdict        VerdictDelta    `json:"verdict"`
	NextAction     NextActionDelta `json:"next_action"`
	TaskCount      int             `json:"task_count_before"`
	TaskCountAfter int             `json:"task_count_after"`
}

type TrustRunDiff struct {
	ReportVersion string          `json:"report_version"`
	Scope         string          `json:"scope"`
	ScopeID       string          `json:"scope_id"`
	Paths         ReportPaths     `json:"paths"`
	Score         ScoreDelta      `json:"score"`
	Verdict       VerdictDelta    `json:"verdict"`
	NextAction    NextActionDelta `json:"next_action"`
}

type DoctorDiff struct {
	ReportVersion string          `json:"report_version"`
	Scope         string          `json:"scope"`
	Paths         ReportPaths     `json:"paths"`
	Ready         BoolDelta       `json:"ready"`
	Warnings      CountDelta      `json:"warnings"`
	Failures      CountDelta      `json:"failures"`
	TrustVerdict  VerdictDelta    `json:"trust_verdict,omitempty"`
	NextAction    NextActionDelta `json:"next_action"`
}

type BoolDelta struct {
	Before  bool `json:"before"`
	After   bool `json:"after"`
	Changed bool `json:"changed"`
}

type CountDelta struct {
	Before  int  `json:"before"`
	After   int  `json:"after"`
	Delta   int  `json:"delta"`
	Changed bool `json:"changed"`
}
