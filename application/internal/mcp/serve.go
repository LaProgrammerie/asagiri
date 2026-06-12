package mcp

import (
	"context"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// ServeStdio runs the MCP server on stdin/stdout.
func ServeStdio(repoRoot string, cfg *config.Config) error {
	s := &Server{RepoRoot: repoRoot, Config: cfg, In: os.Stdin, Out: os.Stdout}
	return s.Serve(context.Background())
}
