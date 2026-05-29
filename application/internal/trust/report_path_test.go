package trust

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteReportRejectsUnsafeTrustID(t *testing.T) {
	repo := t.TempDir()
	_, _, err := WriteReport(repo, "../escape", TrustReport{TrustID: "../escape"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid trust id")
}
