package contextopt

import (
	"testing"
)

func TestReduceTrimsLargeFile(t *testing.T) {
	entries := []FileEntry{{RelPath: "a.go", Content: string(make([]byte, 50000))}}
	out, warns := Reduce(entries, nil, ReduceOpts{MaxCharsPerFile: 1000})
	if len(out[0].Content) > 1500 {
		t.Fatalf("expected trim, got %d", len(out[0].Content))
	}
	if len(warns) != 1 {
		t.Fatalf("warns %v", warns)
	}
}
