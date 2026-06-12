package doctor

const ReportVersion = "doctor-v1"

// Check status values (Checks, Warnings, Failures).
const (
	StatusOK   = "ok"
	StatusWarn = "warn"
	StatusFail = "fail"
)

// Options controls doctor collection.
type Options struct {
	Full bool
}

// Report is the read-only doctor synthesis (JSON-serializable).
//
// Ready is true when there are no blocking failures (StatusFail).
// Warnings do not block Ready; use --strict at the CLI to fail on warnings.
type Report struct {
	ReportVersion    string                   `json:"report_version"`
	Ready            bool                     `json:"ready"`
	Warnings         []Check                  `json:"warnings"`
	Failures         []Check                  `json:"failures"`
	Repository       RepositoryInfo           `json:"repository"`
	State            StateInfo                `json:"state"`
	Gates            []GateInfo               `json:"gates"`
	Agents           []AgentInfo              `json:"agents"`
	AgentRegistry    AgentRegistryInfo        `json:"agent_registry"`
	AgentSpecs       []AgentSpecEntry         `json:"agent_specs"`
	AgentDrift       []AgentDriftEntry        `json:"agent_drift"`
	LastOrchestrated *OrchestratedContextInfo `json:"last_orchestrated,omitempty"`
	MissingTools     []MissingToolInfo        `json:"missing_tools,omitempty"`
	Checks           []Check                  `json:"checks"`
	Trust            *TrustInfo               `json:"trust,omitempty"`
	NextActions      []Action                 `json:"next_actions"`
}

type RepositoryInfo struct {
	GitRoot      string `json:"git_root,omitempty"`
	ConfigPath   string `json:"config_path,omitempty"`
	ConfigLoaded bool   `json:"config_loaded"`
	ConfigError  string `json:"config_error,omitempty"`
}

type StateInfo struct {
	SQLitePath    string `json:"sqlite_path,omitempty"`
	SQLitePresent bool   `json:"sqlite_present"`
	SchemaVersion int    `json:"schema_version,omitempty"`
	RunCount      int    `json:"run_count"`
	TaskCount     int    `json:"task_count"`
	ActiveFeature string `json:"active_feature,omitempty"`
}

type GateInfo struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode,omitempty"`
	Status  string `json:"status"` // active | inactive | disabled | invalid_mode
	Detail  string `json:"detail,omitempty"`
}

type AgentInfo struct {
	Role            string   `json:"role"`
	LogicalID       string   `json:"logical_id,omitempty"`
	Provider        string   `json:"provider,omitempty"`
	Command         string   `json:"command,omitempty"`
	InPath          bool     `json:"in_path"`
	Status          string   `json:"status"` // ok | warn | missing | fail
	Detail          string   `json:"detail,omitempty"`
	SpecVersion     string   `json:"spec_version,omitempty"`
	ContentHash     string   `json:"content_hash,omitempty"`
	StoredHash      string   `json:"stored_hash,omitempty"`
	Drift           []string `json:"drift,omitempty"`
	PromptSource    string   `json:"prompt_source,omitempty"` // disk | embedded | missing | invalid
	OutputFormat    string   `json:"output_format,omitempty"`
	ProviderSupport string   `json:"provider_support,omitempty"`
}

// AgentRegistryInfo describes the on-disk AgentSpec registry.
type AgentRegistryInfo struct {
	Path          string `json:"path"`
	Present       bool   `json:"present"`
	FileCount     int    `json:"file_count"`
	UsingEmbedded bool   `json:"using_embedded_defaults"`
	Status        string `json:"status"`
	Detail        string `json:"detail,omitempty"`
}

// AgentSpecEntry is one config.agents key aligned with AgentSpec / provider runtime.
type AgentSpecEntry struct {
	ConfigKey       string   `json:"config_key"`
	SpecID          string   `json:"spec_id,omitempty"`
	SpecVersion     string   `json:"spec_version,omitempty"`
	Role            string   `json:"role,omitempty"`
	ContentHash     string   `json:"content_hash,omitempty"`
	StoredHash      string   `json:"stored_hash,omitempty"`
	Drift           []string `json:"drift,omitempty"`
	PromptSource    string   `json:"prompt_source"`
	OutputFormat    string   `json:"output_format,omitempty"`
	ProviderType    string   `json:"provider_type,omitempty"`
	ProviderSupport string   `json:"provider_support,omitempty"`
	Status          string   `json:"status"`
	Detail          string   `json:"detail,omitempty"`
}

// AgentDriftEntry is one actionable drift between registry, spec and config.
type AgentDriftEntry struct {
	ConfigKey string `json:"config_key"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
	FixCLI    string `json:"fix_cli,omitempty"`
}

// OrchestratedContextInfo points to the most recent orchestrated agent log context.
type OrchestratedContextInfo struct {
	TaskID    string `json:"task_id,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	Feature   string `json:"feature,omitempty"`
	Phase     string `json:"phase,omitempty"`
	AgentHash string `json:"agent_hash,omitempty"`
	LogPath   string `json:"log_path,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type MissingToolInfo struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
	FixCLI string `json:"fix_cli,omitempty"`
}

type Check struct {
	ID      string `json:"id"`
	Status  string `json:"status"` // ok | warn | fail
	Message string `json:"message,omitempty"`
}

type TrustInfo struct {
	Feature     string `json:"feature,omitempty"`
	Verdict     string `json:"verdict,omitempty"`
	TasksAtRisk int    `json:"tasks_at_risk"`
	Summary     string `json:"summary,omitempty"`
}

type Action struct {
	Title string `json:"title"`
	CLI   string `json:"cli"`
}

// Finalize derives Ready, Warnings and Failures from Checks.
func Finalize(r *Report) {
	r.Warnings = make([]Check, 0)
	r.Failures = make([]Check, 0)
	for _, c := range r.Checks {
		switch c.Status {
		case StatusFail:
			r.Failures = append(r.Failures, c)
		case StatusWarn:
			r.Warnings = append(r.Warnings, c)
		}
	}
	r.Ready = len(r.Failures) == 0
}

// ShouldFail reports whether the CLI should exit non-zero.
func ShouldFail(r Report, strict bool) bool {
	if len(r.Failures) > 0 {
		return true
	}
	return strict && len(r.Warnings) > 0
}
