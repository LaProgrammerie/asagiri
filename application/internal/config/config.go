package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigRel     = ".agentflow/config.yaml"
	DefaultExampleRel    = ".agentflow/config.yaml.example"
	DefaultStateBackend  = "sqlite"
	DefaultStatePath     = ".agentflow/state.sqlite"
	DefaultWorktreesPath = ".agentflow/worktrees"
)

// Config mirrors .agentflow/config.yaml.
type Config struct {
	Project    Project              `yaml:"project"`
	Specs      Specs                `yaml:"specs"`
	State      State                `yaml:"state"`
	Worktrees  Worktrees            `yaml:"worktrees"`
	Agents     map[string]Agent     `yaml:"agents"`
	Validation ValidationConfig     `yaml:"validation"`
	Policies   Policies             `yaml:"policies"`
}

type Project struct {
	Name          string `yaml:"name"`
	DefaultBranch string `yaml:"default_branch"`
}

type Specs struct {
	KiroPath       string `yaml:"kiro_path"`
	ActiveSpecPath string `yaml:"active_spec_path"`
	HandoffPath    string `yaml:"handoff_path"`
}

type State struct {
	Backend string `yaml:"backend"`
	Path    string `yaml:"path"`
}

type Worktrees struct {
	BasePath       string `yaml:"base_path"`
	BranchPrefix   string `yaml:"branch_prefix"`
	CleanupPolicy  string `yaml:"cleanup_policy"`
}

type Agent struct {
	Command        string   `yaml:"command"`
	Args           []string `yaml:"args"`
	Timeout        int      `yaml:"timeout,omitempty"`
	DefaultModel   string   `yaml:"default_model,omitempty"`
	Endpoint       string   `yaml:"endpoint,omitempty"`
	Model          string   `yaml:"model,omitempty"`
	EmbeddingModel string   `yaml:"embedding_model,omitempty"`
}

// ValidationConfig holds named validation commands (spec §7.1).
type ValidationConfig struct {
	Commands []ValidationCommand `yaml:"commands"`
}

// ValidationCommand is one validation step.
type ValidationCommand struct {
	Name     string `yaml:"name"`
	Command  string `yaml:"command"`
	Required bool   `yaml:"required"`
}

// Policies holds safety and governance rules (spec §7.1).
type Policies struct {
	RequireCleanGit              bool     `yaml:"require_clean_git"`
	ForbidUntrackedSecretFiles     bool     `yaml:"forbid_untracked_secret_files"`
	MaxFilesChangedPerTask         int      `yaml:"max_files_changed_per_task"`
	AllowNetwork                   bool     `yaml:"allow_network"`
	RequireHumanApprovalFor        []string `yaml:"require_human_approval_for"`
}

// Load reads and validates config at path relative to repoRoot.
func Load(path, repoRoot string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.applyDefaults(filepath.Base(repoRoot))
	cfg.applyValidationDefaults(repoRoot)
	if err := cfg.Validate(repoRoot); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyValidationDefaults(repoRoot string) {
	if len(c.Validation.Commands) > 0 {
		return
	}
	if cmds := DefaultGoValidationCommandsForRepo(repoRoot); len(cmds) > 0 {
		c.Validation.Commands = cmds
	}
}

func (c *Config) applyDefaults(repoDirName string) {
	if c.Project.Name == "" {
		c.Project.Name = repoDirName
	}
	if c.Project.DefaultBranch == "" {
		c.Project.DefaultBranch = "main"
	}
	if c.Specs.KiroPath == "" {
		c.Specs.KiroPath = ".kiro/specs"
	}
	if c.Specs.ActiveSpecPath == "" {
		c.Specs.ActiveSpecPath = "docs/ai/active/current-spec.md"
	}
	if c.Specs.HandoffPath == "" {
		c.Specs.HandoffPath = "docs/ai/active/handoff.md"
	}
	if c.State.Backend == "" {
		c.State.Backend = DefaultStateBackend
	}
	if c.State.Path == "" {
		c.State.Path = DefaultStatePath
	}
	if c.Worktrees.BasePath == "" {
		c.Worktrees.BasePath = DefaultWorktreesPath
	}
	if c.Worktrees.BranchPrefix == "" {
		c.Worktrees.BranchPrefix = "agentflow"
	}
	if c.Worktrees.CleanupPolicy == "" {
		c.Worktrees.CleanupPolicy = "keep_failed"
	}
	if c.Policies.MaxFilesChangedPerTask == 0 {
		c.Policies.MaxFilesChangedPerTask = 20
	}
}

// DefaultGoValidationCommands returns Go-oriented validation when go.mod exists.
func DefaultGoValidationCommands(repoDirName string) []ValidationCommand {
	_ = repoDirName
	return []ValidationCommand{
		{Name: "tests", Command: "go test ./...", Required: true},
		{Name: "vet", Command: "go vet ./...", Required: true},
		{Name: "lint", Command: "golangci-lint run", Required: false},
	}
}

// DefaultGoValidationCommandsForRepo detects go.mod and returns defaults or nil.
func DefaultGoValidationCommandsForRepo(repoRoot string) []ValidationCommand {
	if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err != nil {
		return nil
	}
	return DefaultGoValidationCommands(filepath.Base(repoRoot))
}

// ValidationCommandLines returns command strings for workflow payloads.
func (c *Config) ValidationCommandLines() []string {
	lines := make([]string, 0, len(c.Validation.Commands))
	for _, cmd := range c.Validation.Commands {
		if cmd.Command != "" {
			lines = append(lines, cmd.Command)
		}
	}
	return lines
}

// Validate checks paths are relative and resolve under repoRoot.
func (c *Config) Validate(repoRoot string) error {
	if c.State.Backend != "sqlite" {
		return fmt.Errorf("state.backend: seul %q est supporté pour l'instant", DefaultStateBackend)
	}

	relPaths := []struct {
		name string
		path string
	}{
		{"specs.kiro_path", c.Specs.KiroPath},
		{"specs.active_spec_path", c.Specs.ActiveSpecPath},
		{"specs.handoff_path", c.Specs.HandoffPath},
		{"state.path", c.State.Path},
		{"worktrees.base_path", c.Worktrees.BasePath},
	}

	for _, p := range relPaths {
		if err := validateRelPath(p.name, p.path, repoRoot); err != nil {
			return err
		}
	}

	return nil
}

func validateRelPath(field, rel, repoRoot string) error {
	if rel == "" {
		return fmt.Errorf("%s: chemin vide", field)
	}
	if filepath.IsAbs(rel) {
		return fmt.Errorf("%s: le chemin doit être relatif au dépôt, pas absolu", field)
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s: le chemin ne doit pas sortir du dépôt", field)
	}
	abs := filepath.Join(repoRoot, clean)
	if !strings.HasPrefix(abs, repoRoot) {
		return fmt.Errorf("%s: le chemin ne doit pas sortir du dépôt", field)
	}
	return nil
}

// Resolve joins repoRoot with a configured relative path.
func (c *Config) Resolve(repoRoot, rel string) string {
	return filepath.Join(repoRoot, filepath.Clean(rel))
}

// StateDBPath returns the absolute path to the SQLite database.
func (c *Config) StateDBPath(repoRoot string) string {
	return c.Resolve(repoRoot, c.State.Path)
}

// ConfigPath returns the default config path under repoRoot.
func ConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, DefaultConfigRel)
}

// ExamplePath returns the example config path under repoRoot.
func ExamplePath(repoRoot string) string {
	return filepath.Join(repoRoot, DefaultExampleRel)
}
