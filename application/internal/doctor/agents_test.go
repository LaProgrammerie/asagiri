package doctor

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/agentsync"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func doctorTestConfig() *config.Config {
	cfg := config.NewTestConfig("doctor-test")
	cfg.Providers = map[string]config.ProviderConfig{
		"exec": {Type: config.ProviderTypeExec, Command: "echo"},
		"bad":  {Type: "unknown-provider-xyz"},
	}
	cfg.Agents = map[string]config.Agent{
		"dev":      {Provider: "exec"},
		"reviewer": {Provider: "exec"},
	}
	cfg.Work.DefaultAgent = "dev"
	cfg.Work.DefaultReviewer = "reviewer"
	return cfg
}

func validDevSpecYAML(hash string) []byte {
	meta := ""
	if hash != "" {
		meta = "\n  content_hash: \"" + hash + "\""
	}
	return []byte(`id: dev
version: "1.0.0"
role: dev
provider_targets:
  - exec
system_prompt: |
  Agent test doctor.
instructions:
  - lire handoff
constraints:
  - scope strict
output_contract:
  format: asagiri-v1
  required_fields:
    - status
metadata:
  labels:
    env: test` + meta + `
`)
}

func writeAgentSpec(t *testing.T, repo, name string, body []byte) {
	t.Helper()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), body, 0o644))
}

func TestCollectAgentSpecsRegistryAbsentOK(t *testing.T) {
	repo := t.TempDir()
	cfg := doctorTestConfig()
	col := collectAgentSpecs(repo, cfg)

	require.False(t, col.registry.Present)
	require.True(t, col.registry.UsingEmbedded)
	require.Equal(t, StatusWarn, col.registry.Status)
	require.Empty(t, col.drifts)
	require.Len(t, col.specs, 2)
	for _, s := range col.specs {
		require.Equal(t, "embedded", s.PromptSource)
		require.Equal(t, StatusOK, s.Status)
	}
	var registryCheck *Check
	for i := range col.checks {
		if col.checks[i].ID == "agent_registry" {
			registryCheck = &col.checks[i]
			break
		}
	}
	require.NotNil(t, registryCheck)
	require.Equal(t, StatusWarn, registryCheck.Status)
	var syncAction *Action
	for i := range col.actions {
		if col.actions[i].CLI == "asa agents sync --write" {
			syncAction = &col.actions[i]
			break
		}
	}
	require.NotNil(t, syncAction)
}

func TestCollectAgentSpecsInvalidSpec(t *testing.T) {
	repo := t.TempDir()
	writeAgentSpec(t, repo, "dev.yaml", []byte("id: [broken"))
	cfg := doctorTestConfig()
	col := collectAgentSpecs(repo, cfg)

	require.True(t, col.registry.Present)
	var invalid *AgentDriftEntry
	for i := range col.drifts {
		if col.drifts[i].Kind == "invalid_spec" {
			invalid = &col.drifts[i]
			break
		}
	}
	require.NotNil(t, invalid)
	require.Equal(t, "dev", invalid.ConfigKey)
}

func TestCollectAgentSpecsConfigAgentWithoutDiskSpec(t *testing.T) {
	repo := t.TempDir()
	writeAgentSpec(t, repo, "reviewer.yaml", validDevSpecYAML(""))
	// reviewer file but id must match filename - fix reviewer spec
	reviewerBody := []byte(`id: reviewer
version: "1.0.0"
role: reviewer
provider_targets:
  - exec
system_prompt: |
  Reviewer.
output_contract:
  format: gate-yaml
`)
	writeAgentSpec(t, repo, "reviewer.yaml", reviewerBody)

	cfg := doctorTestConfig()
	col := collectAgentSpecs(repo, cfg)

	var missing *AgentDriftEntry
	for i := range col.drifts {
		if col.drifts[i].ConfigKey == "dev" && col.drifts[i].Kind == "missing_spec" {
			missing = &col.drifts[i]
			break
		}
	}
	require.NotNil(t, missing)
}

func TestCollectAgentSpecsUnsupportedProvider(t *testing.T) {
	repo := t.TempDir()
	body := []byte(`id: dev
version: "1.0.0"
role: dev
provider_targets:
  - exec
system_prompt: |
  Agent.
output_contract:
  format: asagiri-v1
`)
	writeAgentSpec(t, repo, "dev.yaml", body)
	cfg := doctorTestConfig()
	cfg.Agents["dev"] = config.Agent{Provider: "bad"}
	col := collectAgentSpecs(repo, cfg)

	var unsupported *AgentDriftEntry
	for i := range col.drifts {
		if col.drifts[i].Kind == "unsupported_provider" {
			unsupported = &col.drifts[i]
			break
		}
	}
	require.NotNil(t, unsupported)
}

func TestCollectAgentSpecsHashMismatch(t *testing.T) {
	repo := t.TempDir()
	body := validDevSpecYAML("deadbeef")
	writeAgentSpec(t, repo, "dev.yaml", body)
	cfg := doctorTestConfig()
	col := collectAgentSpecs(repo, cfg)

	var hashDrift *AgentDriftEntry
	for i := range col.drifts {
		if col.drifts[i].Kind == "hash_mismatch" {
			hashDrift = &col.drifts[i]
			break
		}
	}
	require.NotNil(t, hashDrift)
}

func TestFindLastOrchestratedContext(t *testing.T) {
	repo := t.TempDir()
	taskDir := filepath.Join(repo, ".asagiri", "logs", "task-1", "agents", "dev")
	require.NoError(t, os.MkdirAll(taskDir, 0o755))
	ctxPath := filepath.Join(taskDir, "context.json")
	body := []byte(`{"agent_id":"dev","task_id":"task-1","feature":"feat","phase":"dev","agent_hash":"abc"}`)
	require.NoError(t, os.WriteFile(ctxPath, body, 0o644))

	lo := findLastOrchestratedContext(repo)
	require.NotNil(t, lo)
	require.Equal(t, "task-1", lo.TaskID)
	require.Equal(t, "dev", lo.AgentID)
	require.Equal(t, "abc", lo.AgentHash)
}

func TestAgentSpecsJSONStable(t *testing.T) {
	report := Report{
		ReportVersion: ReportVersion,
		AgentRegistry: AgentRegistryInfo{
			Path:      "/repo/.asagiri/agents",
			Present:   true,
			FileCount: 1,
			Status:    StatusOK,
			Detail:    "1 fichier(s) AgentSpec",
		},
		AgentSpecs: []AgentSpecEntry{
			{
				ConfigKey:       "dev",
				SpecID:          "dev",
				SpecVersion:     "1.0.0",
				Role:            "dev",
				ContentHash:     "abc123def456",
				PromptSource:    "disk",
				OutputFormat:    "asagiri-v1",
				ProviderType:    "exec",
				ProviderSupport: "inline_prompt",
				Status:          StatusOK,
			},
		},
		AgentDrift: []AgentDriftEntry{},
		LastOrchestrated: &OrchestratedContextInfo{
			TaskID:  "task-1",
			AgentID: "dev",
			Feature: "feat",
			Phase:   "dev",
			LogPath: ".asagiri/logs/task-1/agents/dev/context.json",
		},
	}
	Finalize(&report)
	var buf bytes.Buffer
	require.NoError(t, FormatJSON(&buf, report))
	assertGolden(t, "doctor_agents.json", buf.String())

	var decoded Report
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	require.Equal(t, "doctor-v1", decoded.ReportVersion)
	require.Len(t, decoded.AgentSpecs, 1)
	require.Equal(t, "dev", decoded.AgentSpecs[0].ConfigKey)
}

func TestFormatTextAgentSpecsSection(t *testing.T) {
	report := Report{
		ReportVersion: ReportVersion,
		Ready:         true,
		Repository:    RepositoryInfo{GitRoot: "/repo", ConfigLoaded: true},
		AgentRegistry: AgentRegistryInfo{
			Path:      "/repo/.asagiri/agents",
			Present:   true,
			FileCount: 2,
			Status:    StatusOK,
		},
		AgentSpecs: []AgentSpecEntry{
			{
				ConfigKey:       "dev",
				SpecVersion:     "1.0.0",
				ContentHash:     "abcdef1234567890",
				PromptSource:    "disk",
				OutputFormat:    "asagiri-v1",
				ProviderType:    "exec",
				ProviderSupport: "inline_prompt",
				Status:          StatusOK,
			},
		},
		AgentDrift: []AgentDriftEntry{
			{ConfigKey: "enricher", Kind: "missing_spec", Message: "agents.enricher configuré sans AgentSpec disque", FixCLI: "asa agents sync --write --agent enricher"},
		},
		LastOrchestrated: &OrchestratedContextInfo{TaskID: "t1", AgentID: "dev", LogPath: ".asagiri/logs/t1/agents/dev/context.json"},
		Checks:           []Check{{ID: "git", Status: StatusOK}},
	}
	Finalize(&report)
	var buf bytes.Buffer
	require.NoError(t, FormatText(&buf, report, false))
	require.Contains(t, buf.String(), "Agent specs")
	require.Contains(t, buf.String(), "Agent drift")
	require.Contains(t, buf.String(), "Dernier contexte")
}

func TestDoctorDriftClearedAfterSyncWrite(t *testing.T) {
	repo := t.TempDir()
	cfg := doctorTestConfig()
	writeAgentSpec(t, repo, "reviewer.yaml", []byte(`id: reviewer
version: "1.0.0"
role: reviewer
system_prompt: test
output_contract:
  format: gate-yaml
`))

	before := collectAgentSpecs(repo, cfg)
	require.NotEmpty(t, before.drifts)

	_, err := agentsync.Sync(repo, agentsync.Options{Write: true})
	require.NoError(t, err)

	after := collectAgentSpecs(repo, cfg)
	require.True(t, after.registry.Present)
	require.GreaterOrEqual(t, after.registry.FileCount, 2)

	for _, d := range after.drifts {
		require.NotEqual(t, "missing_spec", d.Kind, "unexpected drift: %+v", d)
	}
}
