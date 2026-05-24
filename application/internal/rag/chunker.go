package rag

import (
	"strings"
)

const defaultChunkSize = 1200

// SplitText splits text into overlapping chunks for indexing.
func SplitText(path, content string, size int) []TextChunk {
	if size <= 0 {
		size = defaultChunkSize
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	var out []TextChunk
	start := 0
	for start < len(content) {
		end := start + size
		if end > len(content) {
			end = len(content)
		}
		out = append(out, TextChunk{
			Path:    path,
			Offset:  start,
			Content: content[start:end],
		})
		if end == len(content) {
			break
		}
		start = end - size/8
		if start < 0 {
			start = 0
		}
	}
	return out
}

// TextChunk is one indexed text segment.
type TextChunk struct {
	Path    string
	Offset  int
	Content string
}
