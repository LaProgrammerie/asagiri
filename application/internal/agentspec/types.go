package agentspec

// RegistryDir is the agents registry path relative to the repository root.
const RegistryDir = ".asagiri/agents"

// Role constants for AgentSpec.role.
const (
	RoleDev        = "dev"
	RoleReviewer   = "reviewer"
	RoleEnricher   = "enricher"
	RoleGovernance = "governance"
	RoleSpec       = "spec"
	RoleGate       = "gate"
	RoleUtility    = "utility"
)

// Output format constants for output_contract.format.
const (
	OutputAsagiriV1 = "asagiri-v1"
	OutputGateYAML  = "gate-yaml"
	OutputGateJSON  = "gate-json"
	OutputText      = "text"
)

// OutputContract describes the expected agent stdout shape.
type OutputContract struct {
	Format         string   `yaml:"format" json:"format"`
	RequiredFields []string `yaml:"required_fields,omitempty" json:"required_fields,omitempty"`
}

// ExternalSpec holds optional external provider file reference (read-only audit + future sync).
type ExternalSpec struct {
	Kind           string `yaml:"kind,omitempty" json:"kind,omitempty"`
	Path           string `yaml:"path,omitempty" json:"path,omitempty"`
	LastSyncedHash string `yaml:"last_synced_hash,omitempty" json:"last_synced_hash,omitempty"`
}

// Spec is the canonical AgentSpec document (asagiri.dev/v1 flat schema).
type Spec struct {
	ID              string         `yaml:"id" json:"id"`
	Version         string         `yaml:"version" json:"version"`
	Role            string         `yaml:"role" json:"role"`
	ProviderTargets []string       `yaml:"provider_targets,omitempty" json:"provider_targets,omitempty"`
	SystemPrompt    string         `yaml:"system_prompt" json:"system_prompt"`
	Instructions    []string       `yaml:"instructions,omitempty" json:"instructions,omitempty"`
	Constraints     []string       `yaml:"constraints,omitempty" json:"constraints,omitempty"`
	OutputContract  OutputContract `yaml:"output_contract" json:"output_contract"`
	External        *ExternalSpec  `yaml:"external,omitempty" json:"external,omitempty"`
	Metadata        map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Source is the filesystem path or "embedded" — not serialized.
	Source string `yaml:"-" json:"-"`
	// ContentHash is the stable semantic hash — not part of YAML input for hashing.
	ContentHash string `yaml:"-" json:"content_hash,omitempty"`
}

// Meta is a lightweight index entry returned by Loader.List.
type Meta struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Role        string `json:"role"`
	ContentHash string `json:"content_hash"`
	Source      string `json:"source"`
}
