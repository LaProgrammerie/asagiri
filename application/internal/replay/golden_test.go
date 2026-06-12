package replay

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoldenBasicRunManifest(t *testing.T) {
	repo := setupReplayFixture(t, "basic-run")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})
	golden, err := os.ReadFile("testdata/replay/basic-run/expected-artifacts.txt")
	require.NoError(t, err)
	for _, line := range strings.Split(strings.TrimSpace(string(golden)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		require.Contains(t, pkg.Manifest.Artifacts, line)
	}
}

func TestGoldenTrustValidation(t *testing.T) {
	repo := setupReplayFixture(t, "trust-validation")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})
	require.DirExists(t, pkg.Path+"/trust/trust-2026-05-29-trust01")
	require.FileExists(t, pkg.Path+"/trust/trust-2026-05-29-trust01/report.json")
}

func TestGoldenInvestigation(t *testing.T) {
	repo := setupReplayFixture(t, "investigation")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:          repo,
		FromInvestigation: "inv-2026-05-29-test01",
		Config:            testPolicies(),
	})
	require.FileExists(t, pkg.Path+"/investigations/replay-pack.json")
}

func TestProvenanceIndex(t *testing.T) {
	repo := setupReplayFixture(t, "basic-run")
	pkg := mustCreateReplay(t, repo, ReplayCreateRequest{
		RepoRoot:  repo,
		FromGraph: "graph-2026-05-29-test0001",
		Config:    testPolicies(),
	})
	require.FileExists(t, pkg.Path+"/context/provenance.json")
}
