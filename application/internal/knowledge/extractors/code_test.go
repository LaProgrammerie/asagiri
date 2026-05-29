package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	"github.com/stretchr/testify/require"
)

func TestCodeExtractorFileSymbolAndImport(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, "internal", "demo")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	src := `package demo

import "fmt"

type Widget struct{}

func (w *Widget) Run() { fmt.Println("ok") }
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "widget.go"), []byte(src), 0o644))

	nodes, edges, _, err := extractors.CodeExtractor{}.Extract(context.Background(), repo, "")
	require.NoError(t, err)

	ids := nodeIDs(nodes)
	require.Contains(t, ids, "file:internal_demo_widget")
	require.Contains(t, ids, "module:demo")
	require.Contains(t, ids, "symbol:Widget_run")

	var hasDepends bool
	for _, e := range edges {
		if e.Type == knowledge.EdgeTypeDependsOn && e.To == "module:fmt" {
			hasDepends = true
		}
	}
	require.True(t, hasDepends)
}

func TestCodeExtractorKeepsExamplePathFunction(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, "internal", "config")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	src := `package config

func ExamplePath(root string) string { return root }
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "paths.go"), []byte(src), 0o644))

	nodes, _, _, err := extractors.CodeExtractor{}.Extract(context.Background(), repo, "")
	require.NoError(t, err)
	require.Contains(t, nodeIDs(nodes), "symbol:config_ExamplePath")
}

func TestTestExtractorLinksSymbol(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, "internal", "invite")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "service.go"), []byte(`package invite

type InvitationService struct{}
func (s *InvitationService) Invite() error { return nil }
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "service_test.go"), []byte(`package invite

import "testing"

func TestInvitationService(t *testing.T) {}
`), 0o644))

	codeNodes, _, _, err := extractors.CodeExtractor{}.Extract(context.Background(), repo, "")
	require.NoError(t, err)

	_, edges, _, err := extractors.TestExtractor{}.Extract(context.Background(), repo, "")
	require.NoError(t, err)

	ids := nodeIDs(codeNodes)
	require.Contains(t, ids, "symbol:InvitationService_invite")

	var hasTestEdge bool
	for _, e := range edges {
		if e.Type == knowledge.EdgeTypeTests && e.To == "test:InvitationServiceTest" {
			hasTestEdge = true
		}
	}
	require.True(t, hasTestEdge)
}

func TestLinkFlowToCodeInviteMember(t *testing.T) {
	nodes := []knowledge.GraphNode{
		{ID: "action:invite_member", Type: knowledge.NodeTypeAction, Name: "invite_member", Path: "flows/onboarding.flow.yaml"},
		{ID: "symbol:InvitationService_invite", Type: knowledge.NodeTypeSymbol, Name: "InvitationService::Invite", Path: "internal/invitation/invitation_service.go"},
	}
	_, edges, warnings := extractors.LinkFlowToCode(nodes, nil)
	require.Empty(t, warnings)

	var implements bool
	for _, e := range edges {
		if e.Type == knowledge.EdgeTypeImplements && e.From == "action:invite_member" && e.To == "symbol:InvitationService_invite" {
			implements = true
		}
	}
	require.True(t, implements)
}
