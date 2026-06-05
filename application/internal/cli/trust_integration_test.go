package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
)

func copyTrustFixtureProduct(t *testing.T, repo string) {
	t.Helper()
	src := filepath.Join("..", "trust", "checks", "testdata", "minimal-product")
	dest := filepath.Join(repo, ".asagiri", "products", "minimal-product")
	require.NoError(t, copyDirTrust(src, dest))
	graphsSrc := filepath.Join("..", "trust", "checks", "testdata", "graphs-minimal.json")
	analysisDir := filepath.Join(repo, ".asagiri", "analysis", "minimal-product")
	require.NoError(t, os.MkdirAll(analysisDir, 0o755))
	require.NoError(t, copyFileTrust(graphsSrc, filepath.Join(analysisDir, "graphs.json")))
}

func copyFileTrust(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	return err
}

func copyDirTrust(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFileTrust(path, target)
	})
}

func trustConfigYAML() string {
	return `project:
  name: trust-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
verification:
  default_profile: production
  gates:
    production:
      min_confidence:
        overall: 0.0
      required_checks: []
`
}

func TestCLIIntegrationTrustCommands(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/test\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), trustConfigYAML())
	copyTrustFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)

	root.SetArgs([]string{"trust", "gates"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Profile: production")

	output.Reset()
	root.SetArgs([]string{"verify", "trust", "workspace-onboarding", "--product", "minimal-product"})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "Asagiri Trust Engine")

	cfg, err := config.Load(filepath.Join(repo, ".asagiri", "config.yaml"), repo)
	require.NoError(t, err)
	result, err := runTrustVerify(context.Background(), repo, cfg, trust.VerificationRequest{
		Flow:    "workspace-onboarding",
		Product: "minimal-product",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.TrustID)

	jsonOut := new(bytes.Buffer)
	root.SetOut(jsonOut)
	root.SetErr(jsonOut)
	root.SetArgs([]string{"verify", "trust", "workspace-onboarding", "--product", "minimal-product", "--json"})
	require.NoError(t, root.Execute(), jsonOut.String())
	var report trust.TrustReport
	require.NoError(t, json.Unmarshal(jsonOut.Bytes(), &report))
	require.Equal(t, "workspace-onboarding", report.Flow)
	root.SetOut(output)
	root.SetErr(output)

	output.Reset()
	root.SetArgs([]string{"trust", "replay", result.TrustID})
	require.NoError(t, root.Execute(), output.String())
	require.Contains(t, output.String(), "replayed from:")
}

func TestCLIIntegrationTrustCIFailsOnBlockedGate(t *testing.T) {
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "test@example.com")
	runGitCommand(t, repo, "config", "user.name", "Test")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), `project:
  name: trust-ci
state:
  backend: sqlite
  path: .asagiri/state.sqlite
verification:
  gates:
    production:
      min_confidence:
        overall: 1.0
`)
	copyTrustFixtureProduct(t, repo)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	root := newRootCmd()
	output := new(bytes.Buffer)
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"verify", "trust", "workspace-onboarding", "--product", "minimal-product", "--ci"})
	err = root.Execute()
	require.Error(t, err)
	require.ErrorIs(t, err, errTrustCIFailed)
}
