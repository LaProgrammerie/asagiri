package agentcontract_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontract"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/stretchr/testify/require"
)

func devSpec(format string, required ...string) agentspec.Spec {
	return agentspec.Spec{
		ID:   "dev",
		Role: agentspec.RoleDev,
		OutputContract: agentspec.OutputContract{
			Format:         format,
			RequiredFields: required,
		},
	}
}

func TestValidateTextValid(t *testing.T) {
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputText), "Implémentation terminée.")
	require.True(t, res.Valid)
	require.Equal(t, agentspec.OutputText, res.Format)
	require.NotEmpty(t, res.ExtractedSummary)
}

func TestValidateTextEmpty(t *testing.T) {
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputText), "   ")
	require.False(t, res.Valid)
	require.Equal(t, agentcontract.ErrCodeEmptyOutput, res.Errors[0].Code)
}

func TestValidateAsagiriV1Valid(t *testing.T) {
	stdout := `{"status":"completed","summary":"ok","requires_human_review":false}`
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputAsagiriV1, "status", "summary"), stdout)
	require.True(t, res.Valid)
	require.Equal(t, "ok", res.ExtractedSummary)
}

func TestValidateAsagiriV1MissingField(t *testing.T) {
	stdout := `{"status":"completed","requires_human_review":false}`
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputAsagiriV1, "status", "summary"), stdout)
	require.False(t, res.Valid)
	require.Equal(t, agentcontract.ErrCodeMissingRequiredField, res.Errors[0].Code)
	require.Equal(t, "summary", res.Errors[0].Field)
}

func TestValidateAsagiriV1InvalidJSON(t *testing.T) {
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputAsagiriV1, "status"), "not json")
	require.False(t, res.Valid)
	require.Equal(t, agentcontract.ErrCodeInvalidJSON, res.Errors[0].Code)
}

func TestValidateGateYAMLValid(t *testing.T) {
	stdout := `enrich_gate:
  status: pass
  confidence: 0.9
`
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputGateYAML), stdout)
	require.True(t, res.Valid)
	require.Contains(t, res.ExtractedSummary, "status=pass")
}

func TestValidateGateYAMLInvalid(t *testing.T) {
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputGateYAML), "enrich_gate: [broken")
	require.False(t, res.Valid)
	require.Equal(t, agentcontract.ErrCodeInvalidYAML, res.Errors[0].Code)
}

func TestValidateGateJSONValid(t *testing.T) {
	stdout := `{"governance":{"status":"pass","confidence":0.88}}`
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputGateJSON), stdout)
	require.True(t, res.Valid)
}

func TestValidateGateJSONInvalid(t *testing.T) {
	res := agentcontract.ValidateOutput(devSpec(agentspec.OutputGateJSON), `{"governance":`)
	require.False(t, res.Valid)
	require.Equal(t, agentcontract.ErrCodeInvalidJSON, res.Errors[0].Code)
}

func TestValidateUnknownFormat(t *testing.T) {
	res := agentcontract.ValidateOutput(devSpec("custom-format"), "data")
	require.False(t, res.Valid)
	require.Equal(t, agentcontract.ErrCodeUnknownFormat, res.Errors[0].Code)
}

func TestWriteLogStableJSON(t *testing.T) {
	repo := t.TempDir()
	res := agentcontract.ContractValidationResult{
		Valid:  true,
		Format: agentspec.OutputAsagiriV1,
		Errors: []agentcontract.ContractError{},
	}
	require.NoError(t, agentcontract.WriteLog(repo, "task-1", "dev", res))

	path := filepath.Join(repo, ".asagiri", "logs", "task-1", "agents", "dev", "contract.json")
	body, err := os.ReadFile(path)
	require.NoError(t, err)

	goldenPath := filepath.Join("testdata", "contract_valid.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0o755))
		require.NoError(t, os.WriteFile(goldenPath, body, 0o644))
		return
	}
	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(body))

	var decoded agentcontract.ContractValidationResult
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.True(t, decoded.Valid)
}
