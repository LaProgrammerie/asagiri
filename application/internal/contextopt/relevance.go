package contextopt

import (
	"strings"
)

// ScoreByKeywords boosts FileEntry.Score using token overlap with task/feature text.
func ScoreByKeywords(entries []FileEntry, taskText, feature string) {
	keys := tokenize(taskText + " " + feature)
	if len(keys) == 0 {
		return
	}
	for i := range entries {
		hay := strings.ToLower(entries[i].RelPath + "\n" + entries[i].Content)
		var hits float64
		for _, k := range keys {
			if len(k) < 3 {
				continue
			}
			if strings.Contains(hay, k) {
				hits++
			}
		}
		entries[i].Score = hits
	}
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	replacers := strings.NewReplacer("\n", " ", "\t", " ", ",", " ", ".", " ")
	s = replacers.Replace(s)
	parts := strings.Fields(s)
	seen := map[string]struct{}{}
	var out []string
	for _, p := range parts {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
