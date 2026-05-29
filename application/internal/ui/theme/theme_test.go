package theme

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveKnownThemes(t *testing.T) {
	t.Parallel()

	want := []string{"asagiri-dark", "asagiri-light", "minimal", "high-contrast", "cyber"}
	for _, name := range want {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			th, err := Resolve(name)
			require.NoError(t, err)
			require.Equal(t, name, th.Name)
			require.NotEmpty(t, th.Palette.Primary)
			require.NotEmpty(t, th.Palette.Background)
		})
	}
}

func TestResolveEmptyUsesDefault(t *testing.T) {
	t.Parallel()

	th, err := Resolve("")
	require.NoError(t, err)
	require.Equal(t, "asagiri-dark", th.Name)
}

func TestNamesReturnsFiveThemes(t *testing.T) {
	t.Parallel()

	names := Names()
	require.Len(t, names, 5)
	require.Equal(t, []string{
		"asagiri-dark",
		"asagiri-light",
		"cyber",
		"high-contrast",
		"minimal",
	}, names)
}

func TestMustResolveUnknownFallsBackToDefault(t *testing.T) {
	t.Parallel()

	th := MustResolve("unknown")
	require.Equal(t, "asagiri-dark", th.Name)
}

func TestResolveUnknownTheme(t *testing.T) {
	t.Parallel()

	_, err := Resolve("unknown")
	require.Error(t, err)
}

func TestDefaultTheme(t *testing.T) {
	t.Parallel()

	th := Default()
	require.Equal(t, "asagiri-dark", th.Name)
}

func TestResolveCaseInsensitive(t *testing.T) {
	t.Parallel()

	th, err := Resolve("Asagiri-Dark")
	require.NoError(t, err)
	require.Equal(t, "asagiri-dark", th.Name)
}

func TestAllPalettesHaveRequiredTokens(t *testing.T) {
	t.Parallel()

	for _, name := range Names() {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			th, err := Resolve(name)
			require.NoError(t, err)
			p := th.Palette
			require.NotEmpty(t, p.Primary)
			require.NotEmpty(t, p.Muted)
			require.NotEmpty(t, p.Success)
			require.NotEmpty(t, p.Warning)
			require.NotEmpty(t, p.Error)
			require.NotEmpty(t, p.Border)
			require.NotEmpty(t, p.Background)
		})
	}
}

func TestBorderStyleRenders(t *testing.T) {
	t.Parallel()

	styled := Default().BorderStyle().Render("panel")
	require.NotEmpty(t, styled)
	require.Contains(t, styled, "panel")
}
