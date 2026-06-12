package agentslist

const ReportVersion = "agents-list-v1"

// Report is the operator view of AgentSpec entries (JSON-serializable).
type Report struct {
	ReportVersion string       `json:"report_version"`
	Registry      RegistryInfo `json:"registry"`
	Agents        []Entry      `json:"agents"`
}

// RegistryInfo describes the AgentSpec source (disk registry or embedded defaults).
type RegistryInfo struct {
	Path                  string `json:"path"`
	Present               bool   `json:"present"`
	FileCount             int    `json:"file_count"`
	UsingEmbeddedDefaults bool   `json:"using_embedded_defaults"`
}

// Entry is one AgentSpec row for list/show output.
type Entry struct {
	ID              string           `json:"id"`
	Role            string           `json:"role"`
	Version         string           `json:"version"`
	ContentHash     string           `json:"content_hash"`
	StoredHash      string           `json:"stored_hash,omitempty"`
	Source          string           `json:"source"`
	OutputFormat    string           `json:"output_format"`
	ProviderTargets []string         `json:"provider_targets,omitempty"`
	ProviderSupport *ProviderSupport `json:"provider_support,omitempty"`
	Warnings        []string         `json:"warnings,omitempty"`
}

// ProviderSupport summarizes adapter support for configured runtime and declared targets.
type ProviderSupport struct {
	ConfigKey    string            `json:"config_key,omitempty"`
	ProviderType string            `json:"provider_type,omitempty"`
	Level        string            `json:"level,omitempty"`
	Targets      map[string]string `json:"targets,omitempty"`
	Summary      string            `json:"summary,omitempty"`
}
