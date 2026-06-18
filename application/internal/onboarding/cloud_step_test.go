package onboarding_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCloudSkippedDefault(t *testing.T) {
	require.True(t, onboarding.CloudSkipped(""))
	require.True(t, onboarding.CloudSkipped("skip"))
	require.True(t, onboarding.CloudSkipped("local"))
	require.False(t, onboarding.CloudSkipped("connect"))
}

func TestProbeCloudListsProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/me":
			_, _ = w.Write([]byte(`{"id":"u1","email":"dev@test"}`))
		case "/api/v1/projects":
			_, _ = w.Write([]byte(`[{"id":"p1","name":"Demo","slug":"demo"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	result, err := onboarding.ProbeCloud(context.Background(), srv.URL, "tok")
	require.NoError(t, err)
	require.Equal(t, "dev@test", result.Email)
	require.Len(t, result.Projects, 1)
}

func TestApplyFormLocalOnlyKeepsCloudDisabled(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "go.mod"), "module demo\n\ngo 1.25.0\n")

	form := onboarding.Form{
		Step: onboarding.StepReview,
		Answers: onboarding.Answers{
			ProjectName:      "demo",
			DefaultBranch:    "main",
			Stack:            "go",
			DefaultSpecAgent: "kiro",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
			DefaultEnricher:  "ollama",
			FeatureSlug:      "demo-mvp",
			ProductOneLiner:  "Demo",
			CloudMode:        "skip",
		},
	}
	_, err := onboarding.ApplyForm(repo, form)
	require.NoError(t, err)

	cfg := loadConfig(t, repo)
	require.False(t, cfg.Cloud.Enabled)
	require.Empty(t, cfg.Cloud.ProjectID)
}

func TestApplyFormCloudConnectedWritesConfigAndToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/me":
			_, _ = w.Write([]byte(`{"id":"u1","email":"dev@test"}`))
		case "/api/v1/projects":
			_, _ = w.Write([]byte(`[{"id":"p1","name":"Demo","slug":"demo"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "go.mod"), "module demo\n\ngo 1.25.0\n")
	tokenPath := filepath.Join(repo, "token")
	_ = os.Setenv("HOME", repo)
	t.Cleanup(func() { _ = os.Unsetenv("HOME") })

	form := onboarding.Form{
		Step: onboarding.StepReview,
		Answers: onboarding.Answers{
			ProjectName:      "demo",
			DefaultBranch:    "main",
			Stack:            "go",
			DefaultSpecAgent: "kiro",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
			DefaultEnricher:  "ollama",
			FeatureSlug:      "demo-mvp",
			ProductOneLiner:  "Demo",
			CloudMode:        "connect",
			CloudBaseURL:     srv.URL,
			CloudToken:       "secret-tok",
			CloudProjectID:   "p1",
		},
	}
	res, err := onboarding.ApplyForm(repo, form)
	require.NoError(t, err)
	require.True(t, res.Report.Ready)

	cfg := loadConfig(t, repo)
	require.True(t, cfg.Cloud.Enabled)
	require.Equal(t, srv.URL, cfg.Cloud.BaseURL)
	require.Equal(t, "p1", cfg.Cloud.ProjectID)

	report := cloud.BuildStatus(context.Background(), cloud.StatusOptions{
		Config:   cfg,
		Token:    "secret-tok",
		CheckAPI: true,
	})
	require.True(t, report.TokenPresent)
	require.True(t, report.Reachable)
	require.True(t, report.Linked)
	require.Empty(t, report.Error)

	_ = tokenPath
}

func loadConfig(t *testing.T, repo string) *config.Config {
	t.Helper()
	cfg, err := config.Load(config.ConfigPath(repo), repo)
	require.NoError(t, err)
	return cfg
}

func TestMergeCloudPatchOnlyWhenConnect(t *testing.T) {
	cfg := config.NewTestConfig("demo")
	onboarding.MergeCloudPatch(cfg, onboarding.Answers{CloudMode: "skip"})
	require.False(t, cfg.Cloud.Enabled)

	onboarding.MergeCloudPatch(cfg, onboarding.Answers{
		CloudMode:      "connect",
		CloudBaseURL:   "https://cloud.example",
		CloudProjectID: "p1",
	})
	require.True(t, cfg.Cloud.Enabled)
	require.Equal(t, "p1", cfg.Cloud.ProjectID)

	raw, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.Contains(t, string(raw), "project_id: p1")
}
