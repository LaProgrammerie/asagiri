package contextopt

// FileEntry is one collected file with optional text payload.
type FileEntry struct {
	Path     string
	RelPath  string
	Size     int64
	Content  string
	Score    float64
	Language ContentKind
}

// ContentKind hints token estimation ratios.
type ContentKind string

const (
	KindDefault  ContentKind = "default"
	KindCode     ContentKind = "code"
	KindMarkdown ContentKind = "markdown"
	KindJSON     ContentKind = "json"
)

// ContextPack is ordered context for prompting (specv3 §8).
type ContextPack struct {
	TaskObjective      string
	AcceptanceCriteria string
	FileHints          string
	Investigation      string
	FileExcerpts       []FileDigest
	ValidationLines    []string
	OutputFormat       string
}

// FileDigest is a trimmed file snippet for the pack.
type FileDigest struct {
	Path    string
	Excerpt string
}

// OptimizeResult captures before/after token estimates at pack level.
type OptimizeResult struct {
	OriginalTokens  int
	OptimizedTokens int
	SavingsRatio    float64
	Warnings        []string
}
