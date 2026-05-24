package investigation

// InvestigationResult aggregates local fast-path findings (specv3 §9).
type InvestigationResult struct {
	CandidateFiles []string
	RelatedTests   []string
	SensitivePaths []string
	LargeFiles     []string
	GrepHits       []string
	Symbols        []string
	Imports        []string
	Errors         []string
}
