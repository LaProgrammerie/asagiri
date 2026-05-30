package components_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/stretchr/testify/require"
)

func TestReadinessBannerNotReady(t *testing.T) {
	got := components.ReadinessBanner(components.ReadinessBannerViewModel{Ready: false, Score: 72})
	require.Contains(t, got, "72%")
	require.Contains(t, got, "asa onboard")
}

func TestReadinessBannerReadyEmpty(t *testing.T) {
	got := components.ReadinessBanner(components.ReadinessBannerViewModel{Ready: true, Score: 100})
	require.Empty(t, got)
}
