package cloud

const (
	ReportVersionStatus = "cloud-status-v1"
	ReportVersionPush   = "cloud-push-v1"
	ReportVersionLink   = "cloud-link-v1"
)

// MeResponse mirrors GET /api/v1/me.
type MeResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	CreatedAt   string `json:"createdAt"`
}

// Project mirrors a cloud project resource.
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// RunCreateRequest is POST /api/v1/runs body.
type RunCreateRequest struct {
	Project      string  `json:"project"`
	LocalRunID   string  `json:"localRunId"`
	Feature      string  `json:"feature,omitempty"`
	Status       string  `json:"status"`
	StartedAt    string  `json:"startedAt,omitempty"`
	EndedAt      string  `json:"endedAt,omitempty"`
	DurationMs   int64   `json:"durationMs,omitempty"`
	TrustVerdict string  `json:"trustVerdict,omitempty"`
}

// RunResponse is the created run resource.
type RunResponse struct {
	ID         string `json:"id"`
	LocalRunID string `json:"localRunId"`
}

// LedgerCreateRequest is POST /api/v1/ledger-entries body.
type LedgerCreateRequest struct {
	Run           string         `json:"run"`
	AgentID       string         `json:"agentId"`
	Phase         string         `json:"phase,omitempty"`
	StartedAt     string         `json:"startedAt,omitempty"`
	EndedAt       string         `json:"endedAt,omitempty"`
	DurationMs    int64          `json:"durationMs,omitempty"`
	ExitCode      int            `json:"exitCode,omitempty"`
	PromptHash    string         `json:"promptHash,omitempty"`
	ContextHash   string         `json:"contextHash,omitempty"`
	OutputHash    string         `json:"outputHash,omitempty"`
	ContractValid *bool          `json:"contractValid,omitempty"`
	LogDir        string         `json:"logDir,omitempty"`
	RawPayload    map[string]any `json:"rawPayload,omitempty"`
}

// StatusReport is returned by cloud status.
type StatusReport struct {
	ReportVersion string `json:"report_version"`
	Enabled       bool   `json:"enabled"`
	BaseURL       string `json:"base_url"`
	TokenPath     string `json:"token_path"`
	TokenPresent  bool   `json:"token_present"`
	ProjectID     string `json:"project_id,omitempty"`
	Linked        bool   `json:"linked"`
	Reachable     bool   `json:"reachable,omitempty"`
	Me            *MeResponse `json:"me,omitempty"`
	Error         string `json:"error,omitempty"`
}

// PushItem describes one planned or applied push action.
type PushItem struct {
	LocalRunID   string `json:"local_run_id"`
	Feature      string `json:"feature,omitempty"`
	EntryCount   int    `json:"entry_count"`
	RunStatus    string `json:"run_status"`
	DryRun       bool   `json:"dry_run"`
	Skipped      bool   `json:"skipped,omitempty"`
	SkipReason   string `json:"skip_reason,omitempty"`
	CloudRunID   string `json:"cloud_run_id,omitempty"`
	Error        string `json:"error,omitempty"`
}

// PushReport is returned by cloud push.
type PushReport struct {
	ReportVersion string     `json:"report_version"`
	Mode          string     `json:"mode"`
	ProjectID     string     `json:"project_id"`
	BaseURL       string     `json:"base_url"`
	Items         []PushItem `json:"items"`
	Hint          string     `json:"hint,omitempty"`
}

// LinkReport is returned by cloud link.
type LinkReport struct {
	ReportVersion string  `json:"report_version"`
	ProjectID     string  `json:"project_id"`
	ProjectName   string  `json:"project_name,omitempty"`
	ProjectSlug   string  `json:"project_slug,omitempty"`
	ConfigPath    string  `json:"config_path"`
}
