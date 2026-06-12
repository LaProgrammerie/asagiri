package agentexternal

const ReportVersion = "agents-external-v1"

// Status values for ExternalTarget (read-only audit).
const (
	StatusOK             = "ok"
	StatusDrift          = "drift"
	StatusMissingFile    = "missing_file"
	StatusMissingPath    = "missing_path"
	StatusUnsupported    = "unsupported"
	StatusCLIMissing     = "cli_missing"
	StatusPathUnexpanded = "path_unexpanded"
	StatusNotApplicable  = "not_applicable"
	StatusRejected       = "rejected"
)

// ExternalTarget describes one provider external destination (audit only).
type ExternalTarget struct {
	AgentID        string `json:"agent_id"`
	ConfigKey      string `json:"config_key,omitempty"`
	Provider       string `json:"provider"`
	SupportLevel   string `json:"support_level"`
	ExternalKind   string `json:"external_kind,omitempty"`
	PathSource     string `json:"path_source,omitempty"`
	ConfiguredPath string `json:"configured_path,omitempty"`
	DetectedPath   string `json:"detected_path,omitempty"`
	Writable       bool   `json:"writable"`
	InstalledHash  string `json:"installed_hash,omitempty"`
	DesiredHash    string `json:"desired_hash,omitempty"`
	LastSyncedHash string `json:"last_synced_hash,omitempty"`
	CLICommand     string `json:"cli_command,omitempty"`
	CLIAvailable   bool   `json:"cli_available"`
	Status         string `json:"status"`
	Detail         string `json:"detail,omitempty"`
}

// Report is the read-only external sync audit (JSON-serializable).
type Report struct {
	ReportVersion string           `json:"report_version"`
	ReadOnly      bool             `json:"read_only"`
	Policy        string           `json:"policy"`
	Targets       []ExternalTarget `json:"targets"`
	Notes         []string         `json:"notes,omitempty"`
}
