package configcache_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/configcache"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadBundleRoundTrip(t *testing.T) {
	home := t.TempDir()
	store := configcache.NewStore(home)
	fetched := cloud.BundleFetchResult{
		Bundle: cloud.ConfigBundle{
			"projectSlug": "demo",
			"contentHash": "abc",
			"registry": map[string]any{
				"schemaVersion": 1,
			},
		},
		ContentHash:    "abc",
		ReleaseVersion: 1,
		ETag:           "abc",
		RegistrySchema: 1,
	}
	ptr, err := store.SaveBundle("demo", fetched)
	require.NoError(t, err)
	require.Equal(t, 1, ptr.ReleaseVersion)

	bundle, active, meta, ok, err := store.CurrentBundle("demo")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "abc", active.ContentHash)
	require.Equal(t, 1, meta.ReleaseVersion)
	require.Equal(t, float64(1), bundle["registry"].(map[string]any)["schemaVersion"])
}
