package cli

import (
	"bytes"
	"testing"
)

func TestVersion(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected version output")
	}
}

func TestRootHelpShowsWorkflowExample(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	out := buf.String()
	for _, snippet := range []string{
		"Exemple — développer une feature",
		"agentflow init",
		"agentflow plan billing-v2",
		"--dry-run",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(snippet)) {
			t.Fatalf("help missing %q\n%s", snippet, out)
		}
	}
}
