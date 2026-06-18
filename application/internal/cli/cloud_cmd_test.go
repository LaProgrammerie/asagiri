package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloudStatusLogoutLoginSmoke(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	cfgPath := filepath.Join(dir, ".asagiri", "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`project:
  name: test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
worktrees:
  base_path: .asagiri/worktrees
cloud:
  enabled: false
  base_url: http://localhost
  token_path: `+filepath.Join(dir, ".cloud", "token")+`
`), 0o644))

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	tokenDir := filepath.Join(dir, ".tokens")
	require.NoError(t, os.MkdirAll(tokenDir, 0o700))

	root := newRootCmd()

	var statusOut bytes.Buffer
	root.SetOut(&statusOut)
	root.SetErr(&statusOut)
	root.SetArgs([]string{"cloud", "status", "--json"})
	require.NoError(t, root.Execute())

	var status map[string]any
	require.NoError(t, json.Unmarshal(statusOut.Bytes(), &status))
	require.Equal(t, "cloud-status-v1", status["report_version"])
	require.Equal(t, false, status["token_present"])

	var loginOut bytes.Buffer
	root = newRootCmd()
	root.SetOut(&loginOut)
	root.SetErr(&loginOut)
	root.SetArgs([]string{"cloud", "login", "--token", "fake-token", "--base-url", "http://localhost"})
	require.NoError(t, root.Execute())

	var logoutOut bytes.Buffer
	root = newRootCmd()
	root.SetOut(&logoutOut)
	root.SetErr(&logoutOut)
	root.SetArgs([]string{"cloud", "logout", "--json"})
	require.NoError(t, root.Execute())
	requireStdoutSingleJSON(t, logoutOut.Bytes())
}

func TestCloudPushDryRunAll(t *testing.T) {
	dir, _ := cloudPushTestFixture(t)

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	var out bytes.Buffer
	var errOut bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"cloud", "push", "--dry-run", "--all", "--json"})
	require.NoError(t, root.Execute())
	requireStdoutSingleJSON(t, out.Bytes())
	require.Contains(t, string(errOut.Bytes()), "--execute")
}

func TestCloudPushDefaultIsDryRun(t *testing.T) {
	dir, tokenPath := cloudPushTestFixture(t)

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	var out, errOut bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs([]string{"cloud", "push", "--all", "--json"})
	require.NoError(t, root.Execute())
	requireStdoutSingleJSON(t, out.Bytes())
	require.Contains(t, string(errOut.Bytes()), "--execute")

	var report map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "dry-run", report["mode"])
	_ = tokenPath
}

func TestCloudPushExecuteCallsAPI(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	var runPosts, ledgerPosts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/runs":
			runPosts++
			_, _ = w.Write([]byte(`{"id":"cloud-run-1"}`))
		case "/api/v1/ledger-entries":
			ledgerPosts++
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cfgPath := filepath.Join(dir, ".asagiri", "config.yaml")
	tokenPath := filepath.Join(dir, "token")
	require.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	require.NoError(t, os.WriteFile(cfgPath, []byte(`project:
  name: test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
worktrees:
  base_path: .asagiri/worktrees
cloud:
  enabled: true
  base_url: `+srv.URL+`
  project_id: proj-1
  token_path: `+tokenPath+`
`), 0o644))
	require.NoError(t, os.WriteFile(tokenPath, []byte("fake\n"), 0o600))

	ledgerPath := filepath.Join(dir, ".asagiri", "logs", "agents", "ledger.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(ledgerPath), 0o755))
	require.NoError(t, os.WriteFile(ledgerPath, []byte(`{"run_id":"run-1","agent_id":"dev","exit_code":0,"started_at":"2026-01-01T00:00:00Z","ended_at":"2026-01-01T00:01:00Z","duration_ms":1000,"prompt_hash":"a","output_hash":"b","log_dir":"logs"}`+"\n"), 0o644))

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cloud", "push", "--all", "--execute", "--json"})
	require.NoError(t, root.Execute())
	requireStdoutSingleJSON(t, out.Bytes())

	var report map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &report))
	require.Equal(t, "apply", report["mode"])
	require.Equal(t, 1, runPosts)
	require.Equal(t, 1, ledgerPosts)
}

func cloudPushTestFixture(t *testing.T) (dir, tokenPath string) {
	t.Helper()
	dir = t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	cfgPath := filepath.Join(dir, ".asagiri", "config.yaml")

	ledgerPath := filepath.Join(dir, ".asagiri", "logs", "agents", "ledger.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(ledgerPath), 0o755))
	require.NoError(t, os.WriteFile(ledgerPath, []byte(`{"run_id":"run-1","agent_id":"dev","exit_code":0,"started_at":"2026-01-01T00:00:00Z","ended_at":"2026-01-01T00:01:00Z","duration_ms":1000,"prompt_hash":"a","output_hash":"b","log_dir":"logs"}`+"\n"), 0o644))

	tokenPath = filepath.Join(dir, "token")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`project:
  name: test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
worktrees:
  base_path: .asagiri/worktrees
cloud:
  enabled: true
  base_url: http://localhost
  project_id: proj-1
  token_path: `+tokenPath+`
`), 0o644))
	require.NoError(t, os.WriteFile(tokenPath, []byte("fake\n"), 0o600))
	return dir, tokenPath
}
