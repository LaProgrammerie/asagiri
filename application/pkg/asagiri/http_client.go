package asagiri

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// HTTPClient talks to the local runtime REST API (spec-my-A §24.18).
type HTTPClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// HTTPOptions configures the remote runtime client.
type HTTPOptions struct {
	BaseURL string // e.g. http://127.0.0.1:8765
	Token   string
}

// ConnectHTTP returns a client for asa runtime serve.
func ConnectHTTP(opts HTTPOptions) *HTTPClient {
	if opts.BaseURL == "" {
		opts.BaseURL = "http://127.0.0.1:8765"
	}
	return &HTTPClient{
		baseURL: opts.BaseURL,
		token:   opts.Token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("runtime api %s %s: %s", method, path, string(raw))
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}

// Status returns daemon/runtime counters.
func (c *HTTPClient) Status(ctx context.Context) (runtime.DaemonStatus, error) {
	var wrap struct {
		Status runtime.DaemonStatus `json:"status"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/status", nil, &wrap); err != nil {
		return runtime.DaemonStatus{}, err
	}
	return wrap.Status, nil
}

// StartSession creates a session via POST /v1/sessions.
func (c *HTTPClient) StartSession(ctx context.Context, name, productID, flowID string) (runtime.Session, error) {
	var sess runtime.Session
	err := c.do(ctx, http.MethodPost, "/v1/sessions", map[string]string{
		"name": name, "product_id": productID, "flow_id": flowID,
	}, &sess)
	return sess, err
}

// EmitEvent posts a runtime bus event.
func (c *HTTPClient) EmitEvent(ctx context.Context, eventType, sessionID, flowID string, payload map[string]any) (runtime.RuntimeEvent, error) {
	var ev runtime.RuntimeEvent
	err := c.do(ctx, http.MethodPost, "/v1/events", map[string]any{
		"type": eventType, "session_id": sessionID, "flow_id": flowID, "payload": payload,
	}, &ev)
	return ev, err
}

// RunFlow records flow.started and flow.completed via the API.
func (c *HTTPClient) RunFlow(ctx context.Context, sessionID, flowID string) error {
	if _, err := c.EmitEvent(ctx, "flow.started", sessionID, flowID, nil); err != nil {
		return err
	}
	_, err := c.EmitEvent(ctx, "flow.completed", sessionID, flowID, nil)
	return err
}
