package agentexternal

const SyncReportVersion = "agents-external-sync-v1"

const (
	SyncActionCreate   = "create"
	SyncActionUpdate   = "update"
	SyncActionSkip     = "skip"
	SyncActionConflict = "conflict"
	SyncActionReject   = "reject"
)

// SyncOptions controls external provider file sync (dry-run by default).
type SyncOptions struct {
	AgentID   string
	ConfigKey string
	Write     bool
	Force     bool
}

// SyncItem is one agent external sync outcome.
type SyncItem struct {
	AgentID        string `json:"agent_id"`
	ConfigKey      string `json:"config_key,omitempty"`
	Action         string `json:"action"`
	Path           string `json:"path,omitempty"`
	Provider       string `json:"provider,omitempty"`
	SupportLevel   string `json:"support_level,omitempty"`
	ContentHash    string `json:"content_hash,omitempty"`
	InstalledHash  string `json:"installed_hash,omitempty"`
	LastSyncedHash string `json:"last_synced_hash,omitempty"`
	SpecUpdated    bool   `json:"spec_updated,omitempty"`
	Message        string `json:"message,omitempty"`
}

// SyncReport is the external provider sync report (JSON-serializable).
type SyncReport struct {
	ReportVersion string     `json:"report_version"`
	Mode          string     `json:"mode"`
	ReadOnly      bool       `json:"read_only"`
	Wrote         bool       `json:"wrote"`
	Items         []SyncItem `json:"items"`
	Hint          string     `json:"hint,omitempty"`
}
