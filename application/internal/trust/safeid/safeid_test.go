package safeid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAcceptsSafeIDs(t *testing.T) {
	require.NoError(t, Validate("trust-2026-05-29-abc12345"))
}

func TestValidateRejectsPathSegments(t *testing.T) {
	for _, id := range []string{"", "../evil", "a/b", `a\b`, "trust/../x"} {
		require.Error(t, Validate(id), "id=%q", id)
	}
}
