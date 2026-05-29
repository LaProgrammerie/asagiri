package replay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateReplayID(t *testing.T) {
	require.NoError(t, ValidateReplayID("replay-2026-05-29-a1b2c3d4"))
	require.Error(t, ValidateReplayID(""))
	require.Error(t, ValidateReplayID("../bad"))
	require.Error(t, ValidateReplayID("replay-bad-format"))
	require.ErrorIs(t, ValidateReplayID("../bad"), ErrInvalidReplayID)
}

func TestNewReplayID(t *testing.T) {
	id := NewReplayID()
	require.NoError(t, ValidateReplayID(id))
	require.Contains(t, id, "replay-")
}

func TestRedactSecrets(t *testing.T) {
	in := "API_KEY=sk-12345678901234567890123456789012\nBearer abcdef123456\nTOKEN=secretvalue"
	out := RedactSecrets(in)
	require.NotContains(t, out, "secretvalue")
	require.NotContains(t, out, "abcdef123456")
	require.Contains(t, out, "[REDACTED]")
}

func TestShouldRedactFile(t *testing.T) {
	require.True(t, ShouldRedactFile(".env"))
	require.True(t, ShouldRedactFile("credentials.json"))
	require.False(t, ShouldRedactFile("report.md"))
}

func TestCompressLargeFiles(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "prompts")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	path := filepath.Join(sub, "big.txt")
	body := make([]byte, 5000)
	for i := range body {
		body[i] = 'a'
	}
	require.NoError(t, os.WriteFile(path, body, 0o644))

	compressed, err := CompressLargeFiles(dir, []string{"prompts"}, 4096)
	require.NoError(t, err)
	require.Len(t, compressed, 1)
	require.NoFileExists(t, path)
	require.FileExists(t, path+".gz")

	read, err := ReadMaybeCompressed(path)
	require.NoError(t, err)
	require.Len(t, read, len(body))
}

func TestCaptureAndLoadPackage(t *testing.T) {
	repo := setupReplayFixture(t, "basic-run")

	pkg, err := CapturePackage(t.Context(), ReplayCreateRequest{
		RepoRoot:   repo,
		FromGraph:  "graph-2026-05-29-test0001",
		IncludeRuntime: true,
		IncludeEvents:  true,
		Config: CapturePolicies{
			CapturePrompts:       true,
			CaptureRuntimeEvents: true,
			RedactSecrets:        true,
			CompressLargeFiles:   false,
		},
	})
	require.NoError(t, err)
	require.NoError(t, ValidateReplayID(pkg.ID))
	require.FileExists(t, filepath.Join(pkg.Path, ManifestName))
	require.FileExists(t, filepath.Join(pkg.Path, "graph", "execution-graph.json"))

	loaded, err := LoadPackage(repo, pkg.ID)
	require.NoError(t, err)
	require.Equal(t, pkg.ID, loaded.Manifest.ID)
	require.Equal(t, "graph-2026-05-29-test0001", loaded.Manifest.Source.Graph)
}

func TestExecuteReplayOffline(t *testing.T) {
	repo := setupReplayFixture(t, "graph-run")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})

	result, err := ExecuteReplay(t.Context(), ReplayRunRequest{
		RepoRoot: repo,
		ReplayID: pkg.ID,
		Offline:  true,
	})
	require.NoError(t, err)
	require.Equal(t, ModeOffline, result.Mode)
	require.True(t, result.Offline)
}

func TestExecuteReplaySimulation(t *testing.T) {
	repo := setupReplayFixture(t, "graph-run")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})

	result, err := ExecuteReplay(t.Context(), ReplayRunRequest{
		RepoRoot:   repo,
		ReplayID:   pkg.ID,
		Simulation: true,
	})
	require.NoError(t, err)
	require.Equal(t, ModeSimulation, result.Mode)
}

func TestCompareDetectsDivergence(t *testing.T) {
	repo := setupReplayFixture(t, "divergence")
	pkgA := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})

	pkgB := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0002",
		Config:    testPolicies(),
	})

	cmp, err := DefaultComparator().Compare(t.Context(), repo, pkgA.ID, pkgB.ID)
	require.NoError(t, err)
	require.NotEmpty(t, cmp.Differences)
	require.NotEmpty(t, cmp.Divergences)
}

func TestSnapshot(t *testing.T) {
	repo := setupReplayFixture(t, "basic-run")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})

	result, err := DefaultSnapshotter().Snapshot(t.Context(), SnapshotRequest{
		RepoRoot: repo,
		ReplayID: pkg.ID,
		Name:     "before-review",
	})
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(result.Path, ManifestName))
}

func testPolicies() CapturePolicies {
	return CapturePolicies{
		CapturePrompts:         true,
		CaptureRuntimeEvents:   true,
		CaptureAgentOutputs:    true,
		RedactSecrets:          true,
		CompressLargeFiles:     false,
		CompressThresholdBytes: 4096,
	}
}

func TestStrictDetectsModifiedArtifact(t *testing.T) {
	repo := setupReplayFixture(t, "graph-run")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})
	metrics := filepath.Join(pkg.Path, "graph", "metrics.json")
	require.NoError(t, os.WriteFile(metrics, []byte(`{"cost":99}`), 0o644))

	_, err := ExecuteReplay(t.Context(), ReplayRunRequest{
		RepoRoot: repo,
		ReplayID: pkg.ID,
		Offline:  true,
		Strict:   true,
	})
	require.ErrorIs(t, err, ErrStrictDivergence)
}

func TestCaptureAgentOutputs(t *testing.T) {
	repo := setupReplayFixture(t, "graph-run")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})
	require.FileExists(t, filepath.Join(pkg.Path, "reports", "baseline-hashes.json"))
}

func mustCreateReplay(t *testing.T, repo string, req ReplayCreateRequest) ReplayPackage {
	t.Helper()
	if req.Config == (CapturePolicies{}) {
		req.Config = testPolicies()
	}
	pkg, err := CapturePackage(t.Context(), req)
	require.NoError(t, err)
	return pkg
}

func setupReplayFixture(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("testdata", "replay", name)
	repo := t.TempDir()
	copyReplayFixture(t, src, repo)
	return repo
}

func copyReplayFixture(t *testing.T, src, dest string) {
	t.Helper()
	require.NoError(t, filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, body, 0o644)
	}))
}
