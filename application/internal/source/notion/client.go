package notion

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

const defaultAPIBase = "https://api.notion.com/v1"

// Client calls Notion REST API (token never logged).
type Client struct {
	HTTP    *http.Client
	Token   string
	Version string
	BaseURL string
}

func (c *Client) apiBase() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return defaultAPIBase
}

func (c *Client) apiVersion() string {
	if c.Version != "" {
		return c.Version
	}
	return "2022-06-28"
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiBase()+path, rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Notion-Version", c.apiVersion())
	req.Header.Set("Content-Type", "application/json")
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("notion api %s: %s", resp.Status, truncate(string(data), 200))
	}
	return data, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Page represents minimal Notion page payload.
type Page struct {
	ID             string         `json:"id"`
	URL            string         `json:"url"`
	LastEditedTime string         `json:"last_edited_time"`
	Properties     map[string]any `json:"properties"`
}

// GetPage fetches one page by ID.
func (c *Client) GetPage(ctx context.Context, pageID string) (Page, error) {
	pageID = normalizeID(pageID)
	data, err := c.do(ctx, http.MethodGet, "/pages/"+pageID, nil)
	if err != nil {
		return Page{}, err
	}
	var p Page
	if err := json.Unmarshal(data, &p); err != nil {
		return Page{}, err
	}
	return p, nil
}

// QueryDatabase lists pages from a database.
func (c *Client) QueryDatabase(ctx context.Context, databaseID string) ([]Page, error) {
	databaseID = normalizeID(databaseID)
	data, err := c.do(ctx, http.MethodPost, "/databases/"+databaseID+"/query", map[string]any{})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Results []Page `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// ListBlockChildren returns block children for markdown conversion.
func (c *Client) ListBlockChildren(ctx context.Context, blockID string) ([]Block, error) {
	blockID = normalizeID(blockID)
	data, err := c.do(ctx, http.MethodGet, "/blocks/"+blockID+"/children", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Results []Block `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

// Block is a simplified Notion block.
type Block struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// dynamic types
	Paragraph        *RichTextBlock `json:"paragraph,omitempty"`
	Heading1         *RichTextBlock `json:"heading_1,omitempty"`
	Heading2         *RichTextBlock `json:"heading_2,omitempty"`
	Heading3         *RichTextBlock `json:"heading_3,omitempty"`
	BulletedListItem *RichTextBlock `json:"bulleted_list_item,omitempty"`
	ToDo             *ToDoBlock     `json:"to_do,omitempty"`
}

type RichTextBlock struct {
	RichText []RichText `json:"rich_text"`
}

type RichText struct {
	PlainText string `json:"plain_text"`
}

type ToDoBlock struct {
	RichText []RichText `json:"rich_text"`
	Checked  bool       `json:"checked"`
}

// BlocksToMarkdown converts blocks to markdown (spec §8.3).
func BlocksToMarkdown(blocks []Block) (markdown string, tasksYAML string) {
	var md strings.Builder
	var tasks []string
	for _, b := range blocks {
		text := blockText(b)
		if text == "" {
			continue
		}
		switch b.Type {
		case "heading_1":
			md.WriteString("# " + text + "\n\n")
		case "heading_2":
			md.WriteString("## " + text + "\n\n")
		case "heading_3":
			md.WriteString("### " + text + "\n\n")
		case "bulleted_list_item":
			md.WriteString("- " + text + "\n")
		case "to_do":
			checked := ""
			if b.ToDo != nil && b.ToDo.Checked {
				checked = "x"
				tasks = append(tasks, "- [x] "+text)
			} else {
				tasks = append(tasks, "- [ ] "+text)
			}
			fmt.Fprintf(&md, "- [%s] %s\n", checked, text)
		default:
			md.WriteString(text + "\n\n")
			md.WriteString("<!-- unsupported block: " + b.Type + " -->\n\n")
		}
	}
	if len(tasks) > 0 {
		tasksYAML = "tasks:\n"
		for i, t := range tasks {
			tasksYAML += fmt.Sprintf("  - id: task-%03d\n    title: %q\n", i+1, strings.TrimPrefix(t, "- [ ] "))
		}
	}
	return md.String(), tasksYAML
}

func blockText(b Block) string {
	var rt []RichText
	switch {
	case b.Paragraph != nil:
		rt = b.Paragraph.RichText
	case b.Heading1 != nil:
		rt = b.Heading1.RichText
	case b.Heading2 != nil:
		rt = b.Heading2.RichText
	case b.Heading3 != nil:
		rt = b.Heading3.RichText
	case b.BulletedListItem != nil:
		rt = b.BulletedListItem.RichText
	case b.ToDo != nil:
		rt = b.ToDo.RichText
	}
	var parts []string
	for _, r := range rt {
		parts = append(parts, r.PlainText)
	}
	return strings.Join(parts, "")
}

func normalizeID(id string) string {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) == 32 {
		return fmt.Sprintf("%s-%s-%s-%s-%s", id[0:8], id[8:12], id[12:16], id[16:20], id[20:32])
	}
	return id
}

// TitleFromPage extracts title using property name.
func TitleFromPage(p Page, titleProp string) string {
	if titleProp == "" {
		titleProp = "Name"
	}
	prop, ok := p.Properties[titleProp].(map[string]any)
	if !ok {
		return p.ID
	}
	title, ok := prop["title"].([]any)
	if !ok || len(title) == 0 {
		return p.ID
	}
	item, ok := title[0].(map[string]any)
	if !ok {
		return p.ID
	}
	if pt, ok := item["plain_text"].(string); ok {
		return pt
	}
	return p.ID
}

// StatusFromPage reads status select property.
func StatusFromPage(p Page, statusProp string) string {
	if statusProp == "" {
		statusProp = "Status"
	}
	prop, ok := p.Properties[statusProp].(map[string]any)
	if !ok {
		return "draft"
	}
	sel, ok := prop["select"].(map[string]any)
	if !ok {
		return "draft"
	}
	name, _ := sel["name"].(string)
	if name == "" {
		return "draft"
	}
	return strings.ToLower(name)
}
