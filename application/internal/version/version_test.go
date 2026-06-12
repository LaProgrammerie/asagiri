package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultVars(t *testing.T) {
	require.Equal(t, "dev", Version)
	require.Equal(t, "unknown", Commit)
	require.Equal(t, "unknown", Date)
}

func TestStringDefaults(t *testing.T) {
	got := String()
	for _, want := range []string{
		"Asagiri vdev",
		"commit: unknown",
		"built: unknown",
	} {
		require.Contains(t, got, want)
	}
}

func TestStringStructure(t *testing.T) {
	got := String()
	require.False(t, strings.HasSuffix(got, "\n"), "String() must not include trailing newline")

	lines := strings.Split(got, "\n")
	require.Len(t, lines, 3, "spec: Asagiri vX, commit:, built: on three lines")

	require.True(t, strings.HasPrefix(lines[0], "Asagiri v"), "line 1: %q", lines[0])
	require.NotEmpty(t, strings.TrimPrefix(lines[0], "Asagiri v"), "version value after prefix")

	require.True(t, strings.HasPrefix(lines[1], "commit: "), "line 2: %q", lines[1])
	require.NotEmpty(t, strings.TrimSpace(strings.TrimPrefix(lines[1], "commit: ")), "commit value after prefix")

	require.True(t, strings.HasPrefix(lines[2], "built: "), "line 3: %q", lines[2])
	require.NotEmpty(t, strings.TrimSpace(strings.TrimPrefix(lines[2], "built: ")), "built value after prefix")
}
