package theme_test

import (
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

func TestStylesRenderPageHeader(t *testing.T) {
	st := theme.Default().Styles()
	got := st.RenderPageHeader("Project Onboarding Wizard", "Étape Projet · 2/7")
	require.Contains(t, got, "ASAGIRI")
	require.Contains(t, got, "Project Onboarding Wizard")
	require.Equal(t, 1, strings.Count(got, "ASAGIRI"))
}

func TestStylesRenderTabBar(t *testing.T) {
	st := theme.Default().Styles()
	got := st.RenderTabBar([]string{"A", "B", "C"}, 1)
	require.Contains(t, got, "A")
	require.Contains(t, got, "B")
}

func TestStylesRenderProgress(t *testing.T) {
	st := theme.Default().Styles()
	got := st.RenderProgress(75, 100, 10)
	require.Contains(t, got, "75/100")
}
