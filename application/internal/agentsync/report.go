package agentsync

const ReportVersion = "agents-sync-v1"

// Action describes what sync would do or did for one agent.
const (
	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionSkip     = "skip"
	ActionConflict = "conflict"
)

// Options controls agents registry sync (embedded templates → .asagiri/agents).
type Options struct {
	AgentID string
	Write   bool
	Force   bool
}

// Item is one agent sync result.
type Item struct {
	ID      string `json:"id"`
	Action  string `json:"action"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
	Diff    string `json:"diff,omitempty"`
}

// Report is the sync plan or result (JSON-serializable).
type Report struct {
	ReportVersion string `json:"report_version"`
	Mode          string `json:"mode"`
	RegistryDir   string `json:"registry_dir"`
	Wrote         bool   `json:"wrote"`
	Items         []Item `json:"items"`
	Hint          string `json:"hint,omitempty"`
}
