package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func rootWithSplitIO() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	root := newRootCmd()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	return root, &out, &errOut
}

// requireStdoutSingleJSON asserts stdout is exactly one JSON value (jq-compatible).
func requireStdoutSingleJSON(t *testing.T, stdout []byte) {
	t.Helper()
	require.True(t, json.Valid(stdout), "stdout must be valid JSON")
	dec := json.NewDecoder(bytes.NewReader(stdout))
	var raw json.RawMessage
	require.NoError(t, dec.Decode(&raw), "stdout must decode as JSON")
	require.False(t, dec.More(), "stdout must not contain trailing non-JSON content")
}

func requireReportSavedOnStderr(t *testing.T, stderr, wantPath string) {
	t.Helper()
	require.Contains(t, stderr, "Rapport enregistré : "+wantPath)
}
