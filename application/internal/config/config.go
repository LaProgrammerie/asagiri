package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigRel     = ".asagiri/config.yaml"
	DefaultExampleRel    = ".asagiri/config.yaml.example"
	LegacyConfigRel      = ".agentflow/config.yaml"
	LegacyExampleRel     = ".agentflow/config.yaml.example"
	DefaultStateBackend  = "sqlite"
	DefaultStatePath     = ".asagiri/state.sqlite"
	DefaultWorktreesPath = ".asagiri/worktrees"
	DefaultBranchPrefix  = "asagiri"
)

var legacyConfigWarned sync.Map

// Config mirrors .asagiri/config.yaml.
type Config struct {
	Project        Project               `yaml:"project"`
	Specs          Specs                 `yaml:"specs"`
	State          State                 `yaml:"state"`
	Worktrees      Worktrees             `yaml:"worktrees"`
	Agents         map[string]Agent      `yaml:"agents"`
	Validation     ValidationConfig      `yaml:"validation"`
	Policies       Policies              `yaml:"policies"`
	Intent         IntentConfig          `yaml:"intent"`
	Work           WorkConfig            `yaml:"work"`
	Sources        SourcesConfig         `yaml:"sources"`
	Models         ModelsConfig          `yaml:"models"`
	Budgets        BudgetsConfig         `yaml:"budgets"`
	Pricing        PricingConfig         `yaml:"pricing"`
	TokenEst       TokenEstimationConfig `yaml:"token_estimation"`
	Routing        RoutingConfig         `yaml:"routing"`
	UI             UIConfig              `yaml:"ui"`
	MCP            MCPConfig             `yaml:"mcp"`
	Runtime        RuntimeConfig         `yaml:"runtime"`
	Verification   VerificationConfig    `yaml:"verification"`
	ExecutionGraph ExecutionGraphConfig  `yaml:"execution_graph"`
	Coordination   CoordinationConfig    `yaml:"coordination"`
	Knowledge      KnowledgeConfig       `yaml:"knowledge"`
	Replay         ReplayConfig          `yaml:"replay"`
}

// KnowledgeConfig holds engineering knowledge graph defaults (spec-my-E, ADR-024).
type KnowledgeConfig struct {
	DefaultIncludeFlows     bool `yaml:"default_include_flows"`
	DefaultIncludeContracts bool `yaml:"default_include_contracts"`
	DefaultIncludeCode      bool `yaml:"default_include_code"`
	DefaultIncludeTests     bool `yaml:"default_include_tests"`
	IncrementalByDefault    bool `yaml:"incremental_by_default"`
	WarnOnStale             bool `yaml:"warn_on_stale"`
}

// ReplayConfig holds replay capture defaults (spec-my-F §22).
type ReplayConfig struct {
	CapturePrompts         *bool `yaml:"capture_prompts"`
	CaptureRuntimeEvents   *bool `yaml:"capture_runtime_events"`
	CaptureAgentOutputs    *bool `yaml:"capture_agent_outputs"`
	RedactSecrets          *bool `yaml:"redact_secrets"`
	OfflineModeDefault     bool  `yaml:"offline_mode_default"`
	CompressThresholdBytes int   `yaml:"compress_threshold_bytes"`
}

// CoordinationConfig holds multi-agent coordination defaults (spec-my-D §11).
type CoordinationConfig struct {
	MaxParallelAgents        int                            `yaml:"max_parallel_agents"`
	DefaultIsolation         string                         `yaml:"default_isolation"`
	RequireIndependentReview bool                           `yaml:"require_independent_review"`
	AllowSelfReview          bool                           `yaml:"allow_self_review"`
	RequireSecurityReviewFor []string                       `yaml:"require_security_review_for"`
	Assignment               map[string]string              `yaml:"assignment"`
	Profiles                 map[string]CoordinationProfile `yaml:"profiles"`
	Pipeline                 []string                       `yaml:"pipeline"`
	HandoffsPath             string                         `yaml:"handoffs_path"`
	Retry                    CoordinationRetryConfig        `yaml:"retry"`
	Escalation               CoordinationEscalationConfig   `yaml:"escalation"`
	Merge                    CoordinationMergeConfig        `yaml:"merge"`
}

// CoordinationProfile binds a logical profile to an agents: entry (spec-my-D §4).
type CoordinationProfile struct {
	Agent            string   `yaml:"agent"`
	Role             string   `yaml:"role"`
	Capabilities     []string `yaml:"capabilities,omitempty"`
	Restrictions     []string `yaml:"restrictions,omitempty"`
	MaxContextTokens int      `yaml:"max_context_tokens,omitempty"`
	Isolation        string   `yaml:"isolation,omitempty"`
}

// CoordinationRetryConfig holds per-step retry caps (spec-my-D §14).
type CoordinationRetryConfig struct {
	Implementation CoordinationRetryStep `yaml:"implementation"`
}

// CoordinationRetryStep configures retries for one step class.
type CoordinationRetryStep struct {
	MaxAttempts int `yaml:"max_attempts"`
}

// CoordinationEscalationConfig names escalation targets after failures (spec-my-D §14).
type CoordinationEscalationConfig struct {
	AfterFailure       string `yaml:"after_failure"`
	AfterSecondFailure string `yaml:"after_second_failure"`
}

// CoordinationMergeConfig defines merge gates (spec-my-D §16).
type CoordinationMergeConfig struct {
	Require []string `yaml:"require"`
	BlockIf []string `yaml:"block_if"`
}

// ExecutionGraphConfig holds execution graph planner defaults (spec-my-C §24).
type ExecutionGraphConfig struct {
	Enabled                  *bool                  `yaml:"enabled"`
	MaxParallel              int                    `yaml:"max_parallel"`
	DefaultStrategy          string                 `yaml:"default_strategy"`
	RequireCheckpoints       bool                   `yaml:"require_checkpoints"`
	StopOnRisk               string                 `yaml:"stop_on_risk"`
	AllowParallelAgents      bool                   `yaml:"allow_parallel_agents"`
	RequireIsolatedWorktrees bool                   `yaml:"require_isolated_worktrees"`
	Gates                    ExecutionGraphGates    `yaml:"gates"`
	Rollback                 ExecutionGraphRollback `yaml:"rollback"`
}

// IsEnabled reports whether execution graph commands are allowed (default true when unset).
func (c ExecutionGraphConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// ExecutionGraphGates configures trust and approval gates for graph execution.
type ExecutionGraphGates struct {
	TrustRequiredForHighRisk bool     `yaml:"trust_required_for_high_risk"`
	HumanApprovalFor         []string `yaml:"human_approval_for"`
}

// ExecutionGraphRollback configures rollback behaviour for graph runs.
type ExecutionGraphRollback struct {
	RequireStrategyForHighRisk bool `yaml:"require_strategy_for_high_risk"`
	PreserveFailedWorktrees    bool `yaml:"preserve_failed_worktrees"`
}

// VerificationConfig holds trust verification gates (spec-my-B §19).
type VerificationConfig struct {
	DefaultProfile string                 `yaml:"default_profile"`
	Gates          map[string]GateProfile `yaml:"gates"`
}

// GateProfile defines blocking thresholds for a named gate set.
type GateProfile struct {
	MinConfidence  map[string]float64 `yaml:"min_confidence"`
	RequiredChecks []string           `yaml:"required_checks"`
}

// ModelsConfig maps logical model/agent profile ids (specv3 §3.1).
type ModelsConfig map[string]ModelProfile

// ModelProfile describes provider and usage hints for a logical model id.
type ModelProfile struct {
	Provider                    string   `yaml:"provider"`
	Class                       string   `yaml:"class"`
	Model                       string   `yaml:"model"`
	InputCostPer1MTokens        float64  `yaml:"input_cost_per_1m_tokens"`
	OutputCostPer1MTokens       float64  `yaml:"output_cost_per_1m_tokens"`
	TypicalLatencyMsPer1KTokens int      `yaml:"typical_latency_ms_per_1k_tokens"`
	MaxContextTokens            int      `yaml:"max_context_tokens"`
	Usage                       []string `yaml:"usage"`
}

// BudgetsConfig limits estimated spend (specv3 §3.2).
type BudgetsConfig struct {
	DefaultCurrency string         `yaml:"default_currency"`
	PerRun          BudgetLimits   `yaml:"per_run"`
	PerTask         BudgetLimits   `yaml:"per_task"`
	Daily           BudgetLimits   `yaml:"daily"`
	Policies        BudgetPolicies `yaml:"policies"`
}

// BudgetLimits defines numeric caps for one tier.
type BudgetLimits struct {
	MaxEstimatedCost             float64 `yaml:"max_estimated_cost"`
	MaxEstimatedTokens           int     `yaml:"max_estimated_tokens"`
	RequireConfirmationAboveCost float64 `yaml:"require_confirmation_above_cost"`
}

// BudgetPolicies controls blocking and overrides.
type BudgetPolicies struct {
	BlockWhenOverBudget   bool   `yaml:"block_when_over_budget"`
	AllowOverrideWithFlag bool   `yaml:"allow_override_with_flag"`
	OverrideFlag          string `yaml:"override_flag"`
}

// PricingConfig holds per-provider model token prices; no hardcoded defaults (specv3 §6.1).
type PricingConfig struct {
	Currency string                  `yaml:"currency"`
	Models   map[string]ModelPricing `yaml:"models"`
}

// ModelPricing is price sheet entry for a cloud model id (key = model name as used by API).
type ModelPricing struct {
	InputPer1MTokens  float64 `yaml:"input_per_1m_tokens"`
	OutputPer1MTokens float64 `yaml:"output_per_1m_tokens"`
	Source            string  `yaml:"source"`
	UpdatedAt         string  `yaml:"updated_at"`
}

// TokenEstimationConfig tunes chars-per-token heuristics (specv3 §5.3) and provider tokenizers (PF-X-02).
type TokenEstimationConfig struct {
	DefaultCharsPerToken  float64 `yaml:"default_chars_per_token"`
	CodeCharsPerToken     float64 `yaml:"code_chars_per_token"`
	MarkdownCharsPerToken float64 `yaml:"markdown_chars_per_token"`
	JSONCharsPerToken     float64 `yaml:"json_chars_per_token"`
	// DisableProviderTokenizer forces chars-per-token heuristics even when a model is known.
	DisableProviderTokenizer bool `yaml:"disable_provider_tokenizer"`
	// LocalEncoding is the tiktoken encoding name used for local/Ollama models (default cl100k_base).
	LocalEncoding string `yaml:"local_encoding"`
	// AnthropicCharsPerToken tunes the offline Claude ratio when no public tokenizer is available.
	AnthropicCharsPerToken float64 `yaml:"anthropic_chars_per_token"`
	// GoogleCharsPerToken tunes the offline Gemini ratio (no stable Go tokenizer).
	GoogleCharsPerToken float64 `yaml:"google_chars_per_token"`
}

// RoutingConfig selects local vs cloud steps (specv3 §11).
type RoutingConfig struct {
	DefaultStrategy string                     `yaml:"default_strategy"`
	Strategies      map[string]RoutingStrategy `yaml:"strategies"`
}

// RoutingStrategy lists task classes routed to each tier.
type RoutingStrategy struct {
	PreferLocalFor               []string `yaml:"prefer_local_for"`
	UseCloudFastFor              []string `yaml:"use_cloud_fast_for"`
	UseCloudHeavyFor             []string `yaml:"use_cloud_heavy_for"`
	LocalFailuresBeforeCloud     int      `yaml:"local_failures_before_cloud"`
	CloudFastFailuresBeforeHeavy int      `yaml:"cloud_fast_failures_before_heavy"`
}

// UIConfig terminal UX (specv3 §13 + spec-ui §29).
type UIConfig struct {
	Mode                      string `yaml:"mode"` // auto | rich | plain | json
	LiveLogs                  bool   `yaml:"live_logs"`
	ProgressBars              bool   `yaml:"progress_bars"`
	Compact                   bool   `yaml:"compact"`
	DefaultScreen             string `yaml:"default_screen"`
	Theme                     string `yaml:"theme"`
	Mouse                     bool   `yaml:"mouse"`
	Animations                bool   `yaml:"animations"`
	RefreshIntervalMs         int    `yaml:"refresh_interval_ms"`
	CompactThreshold          int    `yaml:"compact_threshold"`
	ShowCLIEquivalents        bool   `yaml:"show_cli_equivalents"`
	ConfirmDestructiveActions bool   `yaml:"confirm_destructive_actions"`
}

// MCPConfig local MCP server limits (specv3 §10).
type MCPConfig struct {
	Enabled            bool                `yaml:"enabled"`
	MaxOutputBytes     int                 `yaml:"max_output_bytes"`
	CommandTimeoutSec  int                 `yaml:"command_timeout_seconds"`
	SecretPathDenylist []string            `yaml:"secret_path_denylist"`
	Investigation      InvestigationConfig `yaml:"investigation"`
}

// InvestigationConfig caps local repo scanning.
type InvestigationConfig struct {
	LargeFileBytes     int64    `yaml:"large_file_bytes"`
	MaxGrepOutputBytes int      `yaml:"max_grep_output_bytes"`
	CommandTimeoutSec  int      `yaml:"command_timeout_seconds"`
	SensitiveGlobs     []string `yaml:"sensitive_globs"`
}

// IntentConfig controls the intent layer (specv2 §9).
type IntentConfig struct {
	Enabled     bool                 `yaml:"enabled"`
	DefaultMode string               `yaml:"default_mode"`
	Resolver    IntentResolverConfig `yaml:"resolver"`
}

// IntentResolverConfig tunes hybrid resolution.
type IntentResolverConfig struct {
	UseOllamaFallback      bool    `yaml:"use_ollama_fallback"`
	MinConfidence          float64 `yaml:"min_confidence"`
	AskWhenBelowConfidence bool    `yaml:"ask_when_below_confidence"`
}

// WorkConfig defaults for work/continue (specv2 §9).
type WorkConfig struct {
	DefaultAgent            string `yaml:"default_agent"`
	DefaultReviewer         string `yaml:"default_reviewer"`
	DefaultEnricher         string `yaml:"default_enricher"`
	StopAfter               string `yaml:"stop_after"`
	AutoVerify              bool   `yaml:"auto_verify"`
	AutoReview              bool   `yaml:"auto_review"`
	MaxTasksPerRun          int    `yaml:"max_tasks_per_run"`
	RequirePlanConfirmation bool   `yaml:"require_plan_confirmation"`
}

// SourcesConfig lists external spec sources (specv2 §9).
type SourcesConfig struct {
	Local  LocalSourceConfig  `yaml:"local"`
	Notion NotionSourceConfig `yaml:"notion"`
}

// LocalSourceConfig scans local spec directories.
type LocalSourceConfig struct {
	Enabled bool     `yaml:"enabled"`
	Paths   []string `yaml:"paths"`
}

// NotionSourceConfig configures Notion sync (specv2 §8).
type NotionSourceConfig struct {
	Enabled             bool   `yaml:"enabled"`
	TokenEnv            string `yaml:"token_env"`
	DefaultDatabaseID   string `yaml:"default_database_id"`
	SpecsDatabaseID     string `yaml:"specs_database_id"`
	TasksDatabaseID     string `yaml:"tasks_database_id"`
	StatusProperty      string `yaml:"status_property"`
	TitleProperty       string `yaml:"title_property"`
	UpdatedTimeProperty string `yaml:"updated_time_property"`
	ImportPath          string `yaml:"import_path"`
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
	BasePath      string `yaml:"base_path"`
	BranchPrefix  string `yaml:"branch_prefix"`
	CleanupPolicy string `yaml:"cleanup_policy"`
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
	RequireCleanGit            bool     `yaml:"require_clean_git"`
	ForbidUntrackedSecretFiles bool     `yaml:"forbid_untracked_secret_files"`
	MaxFilesChangedPerTask     int      `yaml:"max_files_changed_per_task"`
	AllowNetwork               bool     `yaml:"allow_network"`
	RequireHumanApprovalFor    []string `yaml:"require_human_approval_for"`
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
		c.Worktrees.BranchPrefix = DefaultBranchPrefix
	}
	if c.Worktrees.CleanupPolicy == "" {
		c.Worktrees.CleanupPolicy = "keep_failed"
	}
	if c.Policies.MaxFilesChangedPerTask == 0 {
		c.Policies.MaxFilesChangedPerTask = 20
	}
	c.applyIntentDefaults()
	c.applyV3Defaults()
	c.applyExecutionGraphDefaults()
	c.applyCoordinationDefaults()
	c.applyKnowledgeDefaults()
	c.applyReplayDefaults()
}

func (c *Config) applyReplayDefaults() {
	if c.Replay.CompressThresholdBytes == 0 {
		c.Replay.CompressThresholdBytes = 4096
	}
}

func (c *Config) applyKnowledgeDefaults() {
	if c.Knowledge == (KnowledgeConfig{}) {
		c.Knowledge = KnowledgeConfig{
			DefaultIncludeFlows:     true,
			DefaultIncludeContracts: true,
			WarnOnStale:             true,
		}
	}
}

func (c *Config) applyCoordinationDefaults() {
	if c.Coordination.MaxParallelAgents == 0 {
		c.Coordination.MaxParallelAgents = 2
	}
	if c.Coordination.DefaultIsolation == "" {
		c.Coordination.DefaultIsolation = "isolated_worktree"
	}
	if !c.Coordination.RequireIndependentReview {
		c.Coordination.RequireIndependentReview = true
	}
	// AllowSelfReview defaults to false (Go zero value; spec-my-D §11).
	if len(c.Coordination.RequireSecurityReviewFor) == 0 {
		c.Coordination.RequireSecurityReviewFor = []string{
			"auth",
			"permissions",
			"payments",
		}
	}
	if c.Coordination.HandoffsPath == "" {
		c.Coordination.HandoffsPath = ".asagiri/handoffs"
	}
}

func (c *Config) applyExecutionGraphDefaults() {
	if c.ExecutionGraph.MaxParallel == 0 {
		c.ExecutionGraph.MaxParallel = 2
	}
	if c.ExecutionGraph.DefaultStrategy == "" {
		c.ExecutionGraph.DefaultStrategy = "risk_aware"
	}
	if c.ExecutionGraph.StopOnRisk == "" {
		c.ExecutionGraph.StopOnRisk = "high"
	}
	if c.ExecutionGraph.Gates.HumanApprovalFor == nil {
		c.ExecutionGraph.Gates.HumanApprovalFor = []string{
			"migration",
			"security_sensitive",
			"public_contract_change",
		}
	}
}

func (c *Config) applyV3Defaults() {
	hadAnyUISettings := uiHasAnySettings(c.UI)
	if c.Budgets.DefaultCurrency == "" {
		c.Budgets.DefaultCurrency = "EUR"
	}
	if c.Budgets.Policies.OverrideFlag == "" {
		c.Budgets.Policies.OverrideFlag = "--allow-over-budget"
	}
	if c.Pricing.Models == nil {
		c.Pricing.Models = map[string]ModelPricing{}
	}
	if c.Pricing.Currency == "" {
		c.Pricing.Currency = c.Budgets.DefaultCurrency
	}
	if c.TokenEst.DefaultCharsPerToken == 0 {
		c.TokenEst.DefaultCharsPerToken = 4
	}
	if c.TokenEst.CodeCharsPerToken == 0 {
		c.TokenEst.CodeCharsPerToken = 3.2
	}
	if c.TokenEst.MarkdownCharsPerToken == 0 {
		c.TokenEst.MarkdownCharsPerToken = 4.2
	}
	if c.TokenEst.JSONCharsPerToken == 0 {
		c.TokenEst.JSONCharsPerToken = 3.6
	}
	if c.TokenEst.AnthropicCharsPerToken == 0 {
		c.TokenEst.AnthropicCharsPerToken = 3.5
	}
	if c.TokenEst.GoogleCharsPerToken == 0 {
		c.TokenEst.GoogleCharsPerToken = 4.0
	}
	if c.TokenEst.LocalEncoding == "" {
		c.TokenEst.LocalEncoding = "cl100k_base"
	}
	if c.Routing.DefaultStrategy == "" {
		c.Routing.DefaultStrategy = "cost_aware"
	}
	hadUIMode := strings.TrimSpace(c.UI.Mode) != ""
	if c.UI.Mode == "" {
		c.UI.Mode = "auto"
	}
	if !hadUIMode && !c.UI.Compact {
		if !c.UI.LiveLogs {
			c.UI.LiveLogs = true
		}
		if !c.UI.ProgressBars {
			c.UI.ProgressBars = true
		}
	}
	if c.UI.DefaultScreen == "" {
		c.UI.DefaultScreen = "mission"
	}
	if c.UI.Theme == "" {
		c.UI.Theme = "asagiri-dark"
	}
	if c.UI.RefreshIntervalMs == 0 {
		c.UI.RefreshIntervalMs = 500
	}
	if c.UI.CompactThreshold == 0 {
		c.UI.CompactThreshold = 100
	}
	if !hadAnyUISettings {
		c.UI.Mouse = true
		c.UI.Animations = true
		c.UI.ShowCLIEquivalents = true
		c.UI.ConfirmDestructiveActions = true
	}
	if c.MCP.MaxOutputBytes == 0 {
		c.MCP.MaxOutputBytes = 1024 * 1024
	}
	if c.MCP.CommandTimeoutSec == 0 {
		c.MCP.CommandTimeoutSec = 120
	}
	if len(c.MCP.SecretPathDenylist) == 0 {
		c.MCP.SecretPathDenylist = []string{
			".env", ".env.local", ".env.production", "credentials.json", "id_rsa", "id_ed25519",
		}
	}
	if c.MCP.Investigation.LargeFileBytes == 0 {
		c.MCP.Investigation.LargeFileBytes = 512 * 1024
	}
	if c.MCP.Investigation.MaxGrepOutputBytes == 0 {
		c.MCP.Investigation.MaxGrepOutputBytes = 256 * 1024
	}
	if c.MCP.Investigation.CommandTimeoutSec == 0 {
		c.MCP.Investigation.CommandTimeoutSec = c.MCP.CommandTimeoutSec
	}
	if len(c.MCP.Investigation.SensitiveGlobs) == 0 {
		c.MCP.Investigation.SensitiveGlobs = []string{
			"*.pem", "*.key", "*credentials*", "*secret*", ".git/*",
		}
	}
	c.applyRuntimeDefaults()
}

func uiHasAnySettings(ui UIConfig) bool {
	return strings.TrimSpace(ui.Mode) != "" ||
		ui.LiveLogs ||
		ui.ProgressBars ||
		ui.Compact ||
		strings.TrimSpace(ui.DefaultScreen) != "" ||
		strings.TrimSpace(ui.Theme) != "" ||
		ui.Mouse ||
		ui.Animations ||
		ui.RefreshIntervalMs > 0 ||
		ui.CompactThreshold > 0 ||
		ui.ShowCLIEquivalents ||
		ui.ConfirmDestructiveActions
}

// AgentModel returns the configured model id for an agent entry, if any.
func (c *Config) AgentModel(agentName string) string {
	if c == nil || agentName == "" {
		return ""
	}
	a, ok := c.Agents[agentName]
	if !ok {
		return ""
	}
	if strings.TrimSpace(a.Model) != "" {
		return strings.TrimSpace(a.Model)
	}
	return strings.TrimSpace(a.DefaultModel)
}

func (c *Config) applyIntentDefaults() {
	// intent.enabled defaults to true (specv2 §9)
	if c.Intent.DefaultMode == "" && c.Intent.Resolver.MinConfidence == 0 {
		c.Intent.Enabled = true
	} else if c.Intent.DefaultMode != "" && !c.Intent.Enabled {
		c.Intent.Enabled = true
	}
	if c.Intent.DefaultMode == "" {
		c.Intent.DefaultMode = "guided"
	}
	if c.Intent.Resolver.MinConfidence == 0 {
		c.Intent.Resolver.MinConfidence = 0.75
	}
	if !c.Intent.Resolver.UseOllamaFallback {
		c.Intent.Resolver.UseOllamaFallback = true
	}
	if !c.Intent.Resolver.AskWhenBelowConfidence {
		c.Intent.Resolver.AskWhenBelowConfidence = true
	}
	if c.Work.DefaultAgent == "" {
		c.Work.DefaultAgent = "cursor"
	}
	if c.Work.DefaultReviewer == "" {
		c.Work.DefaultReviewer = "codex"
	}
	if c.Work.DefaultEnricher == "" {
		c.Work.DefaultEnricher = "ollama"
	}
	if c.Work.StopAfter == "" {
		c.Work.StopAfter = "report"
	}
	if !c.Work.AutoVerify {
		c.Work.AutoVerify = true
	}
	if c.Work.MaxTasksPerRun == 0 {
		c.Work.MaxTasksPerRun = 1
	}
	if !c.Sources.Local.Enabled && len(c.Sources.Local.Paths) == 0 {
		c.Sources.Local.Enabled = true
	}
	if len(c.Sources.Local.Paths) == 0 {
		c.Sources.Local.Paths = []string{
			".asagiri/specs",
			".kiro/specs",
			"docs/ai/active",
		}
	}
	if c.Sources.Notion.TokenEnv == "" {
		c.Sources.Notion.TokenEnv = "NOTION_TOKEN"
	}
	if c.Sources.Notion.ImportPath == "" {
		c.Sources.Notion.ImportPath = ".asagiri/specs"
	}
	if c.Sources.Notion.StatusProperty == "" {
		c.Sources.Notion.StatusProperty = "Status"
	}
	if c.Sources.Notion.TitleProperty == "" {
		c.Sources.Notion.TitleProperty = "Name"
	}
	if c.Sources.Notion.UpdatedTimeProperty == "" {
		c.Sources.Notion.UpdatedTimeProperty = "Last edited time"
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
		{"coordination.handoffs_path", c.Coordination.HandoffsPath},
	}

	for _, p := range relPaths {
		if err := validateRelPath(p.name, p.path, repoRoot); err != nil {
			return err
		}
	}
	if c.Sources.Notion.ImportPath != "" {
		if err := validateRelPath("sources.notion.import_path", c.Sources.Notion.ImportPath, repoRoot); err != nil {
			return err
		}
	}
	for i, p := range c.Sources.Local.Paths {
		if err := validateRelPath(fmt.Sprintf("sources.local.paths[%d]", i), p, repoRoot); err != nil {
			return err
		}
	}

	if c.MCP.Enabled {
		if c.MCP.MaxOutputBytes <= 0 {
			return fmt.Errorf("mcp.max_output_bytes: doit être > 0 quand mcp.enabled")
		}
		if c.MCP.CommandTimeoutSec <= 0 {
			return fmt.Errorf("mcp.command_timeout_seconds: doit être > 0 quand mcp.enabled")
		}
		if c.MCP.Investigation.CommandTimeoutSec <= 0 {
			return fmt.Errorf("mcp.investigation.command_timeout_seconds: doit être > 0 quand mcp.enabled")
		}
	}
	if c.MCP.MaxOutputBytes < 0 {
		return fmt.Errorf("mcp.max_output_bytes: valeur négative")
	}

	if err := c.validateCoordination(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateCoordination() error {
	co := c.Coordination
	if co.MaxParallelAgents < 1 {
		return fmt.Errorf("coordination.max_parallel_agents: doit être >= 1")
	}
	switch co.DefaultIsolation {
	case "", "shared", "isolated_worktree", "readonly", "sandbox":
	default:
		return fmt.Errorf("coordination.default_isolation: mode inconnu %q", co.DefaultIsolation)
	}
	for id, p := range co.Profiles {
		if strings.TrimSpace(p.Agent) == "" {
			return fmt.Errorf("coordination.profiles[%q].agent: requis", id)
		}
		if _, ok := c.Agents[p.Agent]; !ok {
			return fmt.Errorf("coordination.profiles[%q].agent: %q absent de agents:", id, p.Agent)
		}
		if p.Isolation != "" {
			switch p.Isolation {
			case "shared", "isolated_worktree", "readonly", "sandbox":
			default:
				return fmt.Errorf("coordination.profiles[%q].isolation: mode inconnu %q", id, p.Isolation)
			}
		}
	}
	return nil
}

// NotionToken reads the configured env var for Notion API access.
func (c *Config) NotionToken() string {
	env := c.Sources.Notion.TokenEnv
	if env == "" {
		env = "NOTION_TOKEN"
	}
	return strings.TrimSpace(os.Getenv(env))
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

func warnLegacyConfigDir(dir string) {
	if _, loaded := legacyConfigWarned.LoadOrStore(dir, true); loaded {
		return
	}
	fmt.Fprintf(os.Stderr, "warning: %s is deprecated; migrate to .asagiri/\n", dir)
}

// ResolveConfigPath picks .asagiri/config.yaml, or legacy .agentflow/config.yaml with a warning.
func ResolveConfigPath(repoRoot string) string {
	canon := filepath.Join(repoRoot, DefaultConfigRel)
	if _, err := os.Stat(canon); err == nil {
		return canon
	}
	legacy := filepath.Join(repoRoot, LegacyConfigRel)
	if _, err := os.Stat(legacy); err == nil {
		warnLegacyConfigDir(".agentflow")
		return legacy
	}
	return canon
}

// ConfigPath returns the resolved config path under repoRoot.
func ConfigPath(repoRoot string) string {
	return ResolveConfigPath(repoRoot)
}

// ExamplePath returns the example config path under repoRoot (.asagiri preferred).
func ExamplePath(repoRoot string) string {
	canon := filepath.Join(repoRoot, DefaultExampleRel)
	if _, err := os.Stat(canon); err == nil {
		return canon
	}
	legacy := filepath.Join(repoRoot, LegacyExampleRel)
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return canon
}

// NewTestConfig returns a config with applyDefaults + applyV3Defaults (for unit tests).
func NewTestConfig(repoDirName string) *Config {
	c := &Config{}
	c.applyDefaults(repoDirName)
	c.applyV3Defaults()
	if c.Agents == nil {
		c.Agents = map[string]Agent{}
	}
	if c.Models == nil {
		c.Models = ModelsConfig{}
	}
	if c.Pricing.Models == nil {
		c.Pricing.Models = map[string]ModelPricing{}
	}
	return c
}
