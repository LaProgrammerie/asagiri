package notion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientGetPageMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		if strings.Contains(r.URL.Path, "/pages/") {
			_ = json.NewEncoder(w).Encode(Page{
				ID:             "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
				LastEditedTime: "2026-05-17T10:00:00Z",
				Properties: map[string]any{
					"Name": map[string]any{
						"title": []any{map[string]any{"plain_text": "billing-v2"}},
					},
				},
			})
			return
		}
		if strings.Contains(r.URL.Path, "/blocks/") && strings.Contains(r.URL.Path, "/children") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []Block{{
					Type:     "paragraph",
					Paragraph: &RichTextBlock{RichText: []RichText{{PlainText: "Body"}}},
				}},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := &Client{HTTP: srv.Client(), Token: "test-token", BaseURL: srv.URL}
	page, err := c.GetPage(context.Background(), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.NoError(t, err)
	require.Equal(t, "billing-v2", TitleFromPage(page, "Name"))
	blocks, err := c.ListBlockChildren(context.Background(), page.ID)
	require.NoError(t, err)
	md, _ := BlocksToMarkdown(blocks)
	require.Contains(t, md, "Body")
}

func TestBlocksToMarkdown(t *testing.T) {
	md, tasks := BlocksToMarkdown([]Block{
		{Type: "heading_1", Heading1: &RichTextBlock{RichText: []RichText{{PlainText: "Title"}}}},
		{Type: "to_do", ToDo: &ToDoBlock{RichText: []RichText{{PlainText: "Task one"}}, Checked: false}},
	})
	require.Contains(t, md, "# Title")
	require.Contains(t, tasks, "task-001")
}
