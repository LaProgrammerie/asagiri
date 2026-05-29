package report

import "testing"

func TestCostPerformanceMarkdown(t *testing.T) {
	md := CostPerformanceMarkdown(CostPerformance{
		EstimatedInputTokens:    41800,
		ActualInputTokens:       39950,
		EstimatedOutputTokens:   6000,
		ActualOutputTokens:      5420,
		EstimatedCost:           "€0.08",
		ActualCost:              "€0.07",
		EstimatedDuration:       "2m30s",
		ActualDuration:          "2m12s",
		FilesScanned:            142,
		CandidateFiles:          8,
		LargeFilesSummarized:    3,
		CloudContextReducedFrom: 210000,
		TokenSavingsPercent:     80.1,
	})
	if md == "" {
		t.Fatal("empty markdown")
	}
	for _, want := range []string{"## Cost & Performance", "## Local Work Saved", "80.1%"} {
		if !contains(md, want) {
			t.Fatalf("missing %q in:\n%s", want, md)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
