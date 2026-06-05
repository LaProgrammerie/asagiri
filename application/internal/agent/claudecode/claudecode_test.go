package claudecode_test

import (
	"context"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/agent/claudecode"
)

func TestDefaultConfigHasNonInteractiveFlags(t *testing.T) {
	cfg := claudecode.DefaultConfig()
	if cfg.Command != claudecode.DefaultCommand {
		t.Fatalf("command = %q, want %q", cfg.Command, claudecode.DefaultCommand)
	}
	args := strings.Join(cfg.Args, " ")
	if !strings.Contains(args, "--print") {
		t.Error("DefaultConfig missing --print flag")
	}
	if !strings.Contains(args, "--output-format") {
		t.Error("DefaultConfig missing --output-format flag")
	}
}

func TestAdapterDryRun(t *testing.T) {
	a, err := claudecode.New(claudecode.DefaultConfig(), true /* dryRun */)
	if err != nil {
		t.Fatal(err)
	}
	res, err := a.Run(context.Background(), agent.RunRequest{})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestAdapterName(t *testing.T) {
	a, _ := claudecode.New(claudecode.DefaultConfig(), true)
	if a.Name() != claudecode.AgentName {
		t.Fatalf("Name() = %q, want %q", a.Name(), claudecode.AgentName)
	}
}

func TestAdapterCapabilities(t *testing.T) {
	a, _ := claudecode.New(claudecode.DefaultConfig(), true)
	caps := a.Capabilities()
	if !caps.SupportsJSON {
		t.Error("SupportsJSON should be true for stream-json output")
	}
	if !caps.SupportsFiles {
		t.Error("SupportsFiles should be true")
	}
}
