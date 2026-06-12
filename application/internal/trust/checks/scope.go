package checks

// Scope is the execution scope passed to check runners (mirrors trust.VerificationScope).
type Scope struct {
	TrustID   string
	Flow      string
	Task      string
	Branch    string
	RepoRoot  string
	ProductID string
}
