package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to Asagiri Cloud /api/v1.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// ClientOptions configures the cloud API client.
type ClientOptions struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

// NewClient returns a cloud API client.
func NewClient(opts ClientOptions) *Client {
	base := strings.TrimRight(strings.TrimSpace(opts.BaseURL), "/")
	if base == "" {
		base = "http://localhost"
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		baseURL: base,
		token:   strings.TrimSpace(opts.Token),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) (int, []byte, error) {
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		rdr = bytes.NewReader(raw)
	}
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, rdr)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = resp.Status
		}
		return resp.StatusCode, raw, fmt.Errorf("cloud api %s %s: %s", method, path, RedactError(fmt.Errorf("%s", msg)))
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return resp.StatusCode, raw, fmt.Errorf("cloud api decode: %w", err)
		}
	}
	return resp.StatusCode, raw, nil
}

// Me calls GET /api/v1/me.
func (c *Client) Me(ctx context.Context) (MeResponse, error) {
	var me MeResponse
	_, _, err := c.do(ctx, http.MethodGet, "/api/v1/me", nil, &me)
	return me, err
}

// ListProjects calls GET /api/v1/projects.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	_, raw, err := c.do(ctx, http.MethodGet, "/api/v1/projects", nil, nil)
	if err != nil {
		return nil, err
	}
	return decodeProjectCollection(raw)
}

// CreateRun posts a new run.
func (c *Client) CreateRun(ctx context.Context, req RunCreateRequest) (RunResponse, error) {
	var run RunResponse
	_, _, err := c.do(ctx, http.MethodPost, "/api/v1/runs", req, &run)
	return run, err
}

// CreateLedgerEntry posts a ledger entry.
func (c *Client) CreateLedgerEntry(ctx context.Context, req LedgerCreateRequest) error {
	_, _, err := c.do(ctx, http.MethodPost, "/api/v1/ledger-entries", req, nil)
	return err
}

func decodeProjectCollection(raw []byte) ([]Project, error) {
	var direct []Project
	if err := json.Unmarshal(raw, &direct); err == nil && len(direct) > 0 {
		return direct, nil
	}
	var hydra struct {
		Member []Project `json:"hydra:member"`
	}
	if err := json.Unmarshal(raw, &hydra); err == nil && len(hydra.Member) > 0 {
		return hydra.Member, nil
	}
	var alt struct {
		Member []Project `json:"member"`
	}
	if err := json.Unmarshal(raw, &alt); err == nil {
		return alt.Member, nil
	}
	return []Project{}, nil
}
