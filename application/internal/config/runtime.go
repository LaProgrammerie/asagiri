package config

// RuntimeConfig controls persistent runtime behaviour (spec-my-A §24.17).
type RuntimeConfig struct {
	Mode string           `yaml:"mode"`
	API  RuntimeAPIConfig `yaml:"api"`
}

// RuntimeAPIConfig tunes local embedding API.
type RuntimeAPIConfig struct {
	Port   int    `yaml:"port"`
	Socket string `yaml:"socket"`
}

const (
	RuntimeModeGuided      = "guided"
	RuntimeModeInteractive = "interactive"
	RuntimeModeHeadless    = "headless"
	RuntimeModeCI          = "ci"
	RuntimeModeReview      = "review"
	RuntimeModeExploration = "exploration"
)

// ValidRuntimeModes lists accepted runtime.mode values.
var ValidRuntimeModes = []string{
	RuntimeModeGuided,
	RuntimeModeInteractive,
	RuntimeModeHeadless,
	RuntimeModeCI,
	RuntimeModeReview,
	RuntimeModeExploration,
}

func (c *Config) applyRuntimeDefaults() {
	if c.Runtime.Mode == "" {
		c.Runtime.Mode = RuntimeModeGuided
	}
	if c.Runtime.API.Port == 0 {
		c.Runtime.API.Port = 8765
	}
}
