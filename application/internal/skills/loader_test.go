package skills_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/skills"
	"github.com/stretchr/testify/require"
)

func TestLoadAll(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	root := filepath.Join(dir, ".asagiri", "skills", "review")
	require.NoError(t, os.MkdirAll(root, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "api-review.yaml"), []byte(`---
id: api-platform-review
name: API Platform Review
scope: [backend, architecture]
capabilities: [contract_review]
`), 0o644))

	all, err := skills.LoadAll(dir)
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.Equal(t, "api-platform-review", all[0].ID)

	matched := skills.Match(all, []string{"backend"})
	require.Len(t, matched, 1)
}
