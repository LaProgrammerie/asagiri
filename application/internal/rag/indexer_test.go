package rag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexFixtureRepo(t *testing.T) {
	repo := t.TempDir()
	src := filepath.Join(repo, "application")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Index(IndexOptions{RepoRoot: repo, Paths: []string{"application"}, DryRun: false})
	if err != nil {
		t.Fatal(err)
	}
	if res.Files < 1 || res.Chunks < 1 {
		t.Fatalf("files=%d chunks=%d", res.Files, res.Chunks)
	}
}
