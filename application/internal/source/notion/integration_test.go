//go:build integration

package notion

import (
	"context"
	"os"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNotionIntegrationOptional(t *testing.T) {
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		t.Skip("NOTION_TOKEN not set")
	}
	pageID := os.Getenv("NOTION_TEST_PAGE_ID")
	if pageID == "" {
		t.Skip("NOTION_TEST_PAGE_ID not set")
	}
	c := &Client{Token: token}
	_, err := c.GetPage(context.Background(), pageID)
	require.NoError(t, err)
	_ = config.NotionSourceConfig{}
}
