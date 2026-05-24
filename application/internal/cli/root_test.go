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
