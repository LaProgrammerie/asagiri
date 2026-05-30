package detect_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/onboarding/detect"
	"github.com/stretchr/testify/require"
)

func fixtureRoot(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(file), "..", "fixtures", name)
}

func TestGoDetector(t *testing.T) {
	root := fixtureRoot(t, "go")
	match, err := detect.GoDetector{}.Detect(root)
	require.NoError(t, err)
	require.Equal(t, "go", match.ID)
	cmds := detect.GoDetector{}.ValidationCommands(root)
	require.NotEmpty(t, cmds)
	require.Contains(t, cmds[0].Command, "go test")
}

func TestCastorDetector(t *testing.T) {
	root := fixtureRoot(t, "castor")
	match, err := detect.CastorDetector{}.Detect(root)
	require.NoError(t, err)
	require.Equal(t, "castor", match.ID)
	cmds := detect.CastorDetector{}.ValidationCommands(root)
	require.Len(t, cmds, 2)
	require.Equal(t, "castor qa:static-checks", cmds[0].Command)
}

func TestNodeDetector(t *testing.T) {
	root := fixtureRoot(t, "node")
	match, err := detect.NodeDetector{}.Detect(root)
	require.NoError(t, err)
	require.Equal(t, "node", match.ID)
	cmds := detect.NodeDetector{}.ValidationCommands(root)
	require.NotEmpty(t, cmds)
	require.Equal(t, "npm run qa:js", cmds[0].Command)
}

func TestDetectAllMerge(t *testing.T) {
	root := fixtureRoot(t, "go")
	matches, cmds := detect.DetectAll(root, "auto")
	require.NotEmpty(t, matches)
	require.NotEmpty(t, cmds)
}

func TestDetectAllOverridePHP(t *testing.T) {
	root := fixtureRoot(t, "castor")
	_, cmds := detect.DetectAll(root, "php")
	require.Len(t, cmds, 2)
}
