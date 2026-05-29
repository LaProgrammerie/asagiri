package config

// RuntimeConfig controls persistent runtime behaviour (spec-my-A §24.17).
type RuntimeConfig struct {
	Mode   string              `yaml:"mode"`
	API    RuntimeAPIConfig    `yaml:"api"`
	Memory RuntimeMemoryConfig `yaml:"memory"`
}

// RuntimeMemoryConfig selects the memory embedding backend (spec-phase-finale PF-A-01).
type RuntimeMemoryConfig struct {
	Embedder string                    `yaml:"embedder"`
	Ollama   RuntimeMemoryOllamaConfig `yaml:"ollama"`
	Cloud    RuntimeMemoryCloudConfig  `yaml:"cloud"`
}

// RuntimeMemoryOllamaConfig tunes local Ollama embeddings.
type RuntimeMemoryOllamaConfig struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

// RuntimeMemoryCloudConfig is opt-in OpenAI-compatible embeddings.
type RuntimeMemoryCloudConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url"`
	Model    string `yaml:"model"`
	TokenEnv string `yaml:"token_env"`
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
	if c.Runtime.Memory.Embedder == "" {
		c.Runtime.Memory.Embedder = "hash"
	}
	if c.Runtime.Memory.Ollama.BaseURL == "" {
		c.Runtime.Memory.Ollama.BaseURL = "http://127.0.0.1:11434"
	}
	if c.Runtime.Memory.Ollama.Model == "" {
		c.Runtime.Memory.Ollama.Model = "nomic-embed-text"
	}
	if c.Runtime.Memory.Cloud.TokenEnv == "" {
		c.Runtime.Memory.Cloud.TokenEnv = "OPENAI_API_KEY"
	}
	if c.Runtime.Memory.Cloud.Model == "" {
		c.Runtime.Memory.Cloud.Model = "text-embedding-3-small"
	}
}
