package gates

import "strings"

// ClassifyResult applies finding-severity rules and fail_on filters to produce the final verdict.
func ClassifyResult(r Result, failOn []string) Result {
	if strings.TrimSpace(r.ParseError) != "" {
		r.Status = VerdictFail
		return r
	}
	blocking := failOnSet(failOn)
	hasWarn := false
	for _, f := range r.Findings {
		switch f.Severity {
		case "fail":
			code := strings.ToLower(strings.TrimSpace(f.Code))
			if len(blocking) == 0 || blocking[code] || code == "" {
				r.Status = VerdictFail
				return r
			}
		case "warn":
			hasWarn = true
		}
	}
	if hasWarn {
		r.Status = VerdictWarn
		return r
	}
	switch r.Status {
	case VerdictFail, VerdictWarn, VerdictPass:
		return r
	default:
		r.Status = VerdictPass
		return r
	}
}

func failOnSet(codes []string) map[string]bool {
	if len(codes) == 0 {
		return nil
	}
	out := make(map[string]bool, len(codes))
	for _, c := range codes {
		c = strings.ToLower(strings.TrimSpace(c))
		if c != "" {
			out[c] = true
		}
	}
	return out
}
