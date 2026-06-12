package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime/api"
	"github.com/stretchr/testify/require"
)

func TestRuntimeAPIStatusAndSession(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	store, err := runtime.Open(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	token := "test-token"
	handler := api.NewServer(store).Handler()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Asagiri-Token") != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	reqStatus, err := http.NewRequest(http.MethodGet, srv.URL+"/v1/status", nil)
	require.NoError(t, err)
	reqStatus.Header.Set("X-Asagiri-Token", token)
	resp, err := http.DefaultClient.Do(reqStatus)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	body, _ := json.Marshal(map[string]string{"name": "api-session", "product_id": "p1"})
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/sessions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Asagiri-Token", token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	var sess map[string]any
	require.NoError(t, json.Unmarshal(raw, &sess))
	require.NotEmpty(t, sess["id"])
}

func TestServeBindsLocalhost(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- api.Serve(ctx, api.Options{RepoRoot: repo, Port: 18765})
	}()
	time.Sleep(100 * time.Millisecond)
	resp, err := http.Get("http://127.0.0.1:18765/v1/status")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
	cancel()
	<-errCh
}
