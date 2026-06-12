package doctor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFinalizeReadyWarningsFailures(t *testing.T) {
	r := Report{
		Checks: []Check{
			{ID: "git", Status: StatusOK},
			{ID: "docs.product", Status: StatusWarn, Message: "placeholder"},
			{ID: "sqlite", Status: StatusFail, Message: "absent"},
		},
	}
	Finalize(&r)
	require.False(t, r.Ready)
	require.Len(t, r.Failures, 1)
	require.Equal(t, "sqlite", r.Failures[0].ID)
	require.Len(t, r.Warnings, 1)
	require.Equal(t, "docs.product", r.Warnings[0].ID)
}

func TestShouldFailStrict(t *testing.T) {
	warnOnly := Report{Ready: true, Warnings: []Check{{ID: "w", Status: StatusWarn}}}
	require.False(t, ShouldFail(warnOnly, false))
	require.True(t, ShouldFail(warnOnly, true))

	fail := Report{Ready: false, Failures: []Check{{ID: "f", Status: StatusFail}}}
	require.True(t, ShouldFail(fail, false))
	require.True(t, ShouldFail(fail, true))

	ok := Report{Ready: true}
	require.False(t, ShouldFail(ok, false))
	require.False(t, ShouldFail(ok, true))
}
