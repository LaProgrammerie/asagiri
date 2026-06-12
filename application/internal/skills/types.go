package skills

// Skill is a reusable capability pack (spec-my-A §24.15).
type Skill struct {
	ID           string   `yaml:"id" json:"id"`
	Name         string   `yaml:"name" json:"name"`
	Scope        []string `yaml:"scope" json:"scope,omitempty"`
	Capabilities []string `yaml:"capabilities" json:"capabilities,omitempty"`
	Rules        []string `yaml:"rules" json:"rules,omitempty"`
	Checks       []string `yaml:"checks" json:"checks,omitempty"`
	Metrics      []string `yaml:"metrics" json:"metrics,omitempty"`
	Path         string   `yaml:"-" json:"path,omitempty"`
}
