package agentspec_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/stretchr/testify/require"
)

func TestValidateRejectsIDFilenameMismatch(t *testing.T) {
	spec := agentspec.Spec{
		ID:           "dev",
		Version:      "1.0.0",
		Role:         agentspec.RoleDev,
		SystemPrompt: "ok",
		OutputContract: agentspec.OutputContract{
			Format: agentspec.OutputAsagiriV1,
		},
	}
	err := agentspec.Validate(spec, "reviewer.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ne correspond pas")
}
