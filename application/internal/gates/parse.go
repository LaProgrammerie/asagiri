package gates

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseConfig tunes block extraction for a specific gate type.
type ParseConfig struct {
	BlockKey          string
	MissingBlockError string
	ParseErrorNote    string
}

func (c ParseConfig) withDefaults() ParseConfig {
	out := c
	if strings.TrimSpace(out.BlockKey) == "" {
		out.BlockKey = "gate"
	}
	if strings.TrimSpace(out.MissingBlockError) == "" {
		out.MissingBlockError = out.BlockKey + " block missing from agent output"
	}
	if strings.TrimSpace(out.ParseErrorNote) == "" {
		out.ParseErrorNote = "gate_parse_error"
	}
	return out
}

// ParseResult extracts and normalizes a gate verdict from agent stdout (YAML or JSON).
func ParseResult(stdout string, cfg ParseConfig) Result {
	cfg = cfg.withDefaults()
	raw := extractPayload(stdout, cfg.BlockKey)
	if raw == "" {
		return Result{
			Status:     VerdictFail,
			ParseError: cfg.MissingBlockError,
			Notes:      []string{cfg.ParseErrorNote},
		}
	}

	var wrapper map[string]Result
	if err := yaml.Unmarshal([]byte(raw), &wrapper); err != nil {
		if jsonErr := json.Unmarshal([]byte(raw), &wrapper); jsonErr != nil {
			return Result{
				Status:     VerdictFail,
				ParseError: fmt.Sprintf("%s parse error: %v", cfg.BlockKey, err),
				Notes:      []string{cfg.ParseErrorNote},
			}
		}
	}
	block, ok := wrapper[cfg.BlockKey]
	if !ok {
		// Agent may return the inner block without the wrapper key when raw already starts at blockKey:
		var inner Result
		if err := yaml.Unmarshal([]byte(raw), &inner); err != nil {
			if jsonErr := json.Unmarshal([]byte(raw), &inner); jsonErr != nil {
				return Result{
					Status:     VerdictFail,
					ParseError: fmt.Sprintf("%s parse error: %v", cfg.BlockKey, err),
					Notes:      []string{cfg.ParseErrorNote},
				}
			}
		}
		block = inner
	}

	block.Status = normalizeStatus(string(block.Status))
	block.Confidence = clampConfidence(block.Confidence)
	for i := range block.Findings {
		block.Findings[i].Severity = normalizeSeverity(block.Findings[i].Severity)
		block.Findings[i].Code = strings.TrimSpace(block.Findings[i].Code)
	}
	return block
}

func extractPayload(stdout, blockKey string) string {
	s := strings.TrimSpace(stdout)
	if s == "" {
		return ""
	}
	if fenced := extractYAMLFence(s); fenced != "" {
		s = fenced
	}
	needle := strings.ToLower(blockKey) + ":"
	idx := strings.Index(strings.ToLower(s), needle)
	if idx >= 0 {
		return strings.TrimSpace(s[idx:])
	}
	return s
}

func extractYAMLFence(s string) string {
	lower := strings.ToLower(s)
	start := strings.Index(lower, "```yaml")
	if start < 0 {
		start = strings.Index(lower, "```yml")
	}
	if start < 0 {
		return ""
	}
	rest := s[start:]
	if i := strings.Index(rest, "\n"); i >= 0 {
		rest = rest[i+1:]
	}
	end := strings.Index(rest, "```")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func normalizeStatus(s string) Verdict {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pass", "passed", "ok":
		return VerdictPass
	case "warn", "warning":
		return VerdictWarn
	case "fail", "failed", "error":
		return VerdictFail
	default:
		return Verdict(strings.ToLower(strings.TrimSpace(s)))
	}
}

func normalizeSeverity(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "fail", "failed", "error", "block", "blocking":
		return "fail"
	case "warn", "warning", "advisory":
		return "warn"
	default:
		return strings.ToLower(strings.TrimSpace(s))
	}
}

func clampConfidence(c float64) float64 {
	if c < 0 {
		return 0
	}
	if c > 1 {
		return 1
	}
	return c
}
