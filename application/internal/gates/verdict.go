package gates

// Verdict is the classified outcome of a gate evaluation.
type Verdict string

const (
	VerdictPass Verdict = "pass"
	VerdictWarn Verdict = "warn"
	VerdictFail Verdict = "fail"
)

func (v Verdict) String() string {
	return string(v)
}
