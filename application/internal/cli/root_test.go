package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/version"
	"github.com/stretchr/testify/require"
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

	out := strings.TrimSpace(buf.String())
	want := strings.TrimSpace(version.String())
	if out != want {
		t.Fatalf("version output:\n%q\nwant:\n%q", out, want)
	}
}

func TestVersionFormat(t *testing.T) {
	got := version.String()
	for _, line := range []string{"Asagiri v", "commit: ", "built: "} {
		require.Contains(t, got, line)
	}
}

func TestVersionCLIOutputStructure(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"version"})

	require.NoError(t, root.Execute())

	raw := buf.String()
	require.True(t, strings.HasSuffix(raw, "\n"), "Fprintln must end output with newline")
	body := strings.TrimSuffix(raw, "\n")

	lines := strings.Split(body, "\n")
	require.Len(t, lines, 3, "spec: three lines before trailing newline")

	require.Equal(t, version.String(), body)
	require.True(t, strings.HasPrefix(lines[0], "Asagiri v"))
	require.True(t, strings.HasPrefix(lines[1], "commit: "))
	require.True(t, strings.HasPrefix(lines[2], "built: "))
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
		"asa init",
		"asa plan billing-v2",
		"--dry-run",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(snippet)) {
			t.Fatalf("help missing %q\n%s", snippet, out)
		}
	}
}
