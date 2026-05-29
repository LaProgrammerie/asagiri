package replay

import (
	"regexp"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// CapturePolicies controls what is captured and how secrets are handled (spec §22-23).
type CapturePolicies struct {
	CapturePrompts        bool
	CaptureRuntimeEvents  bool
	CaptureAgentOutputs   bool
	RedactSecrets         bool
	OfflineModeDefault    bool
	CompressLargeFiles    bool
	CompressThresholdBytes int64
}

// DefaultCapturePolicies returns safe defaults from config.
func DefaultCapturePolicies(cfg *config.Config) CapturePolicies {
	p := CapturePolicies{
		CapturePrompts:         true,
		CaptureRuntimeEvents:   true,
		CaptureAgentOutputs:    true,
		RedactSecrets:          true,
		CompressLargeFiles:     true,
		CompressThresholdBytes: 4096,
	}
	if cfg == nil {
		return p
	}
	r := cfg.Replay
	if r.CapturePrompts != nil {
		p.CapturePrompts = *r.CapturePrompts
	}
	if r.CaptureRuntimeEvents != nil {
		p.CaptureRuntimeEvents = *r.CaptureRuntimeEvents
	}
	if r.CaptureAgentOutputs != nil {
		p.CaptureAgentOutputs = *r.CaptureAgentOutputs
	}
	if r.RedactSecrets != nil {
		p.RedactSecrets = *r.RedactSecrets
	}
	if r.OfflineModeDefault {
		p.OfflineModeDefault = true
	}
	if r.CompressThresholdBytes > 0 {
		p.CompressThresholdBytes = int64(r.CompressThresholdBytes)
	}
	return p
}

var (
	tokenPattern    = regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password|credential|auth)[\s:=]+[^\s'"]+`)
	bearerPattern   = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-._~+/]+=*`)
	envLinePattern  = regexp.MustCompile(`(?m)^([A-Z0-9_]+(?:KEY|TOKEN|SECRET|PASSWORD|CREDENTIAL)[A-Z0-9_]*)=.*$`)
	skPattern       = regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`)
)

const redactedPlaceholder = "[REDACTED]"

// RedactSecrets scrubs sensitive patterns from text (spec §23).
func RedactSecrets(text string) string {
	if text == "" {
		return text
	}
	out := text
	out = tokenPattern.ReplaceAllString(out, "${1}="+redactedPlaceholder)
	out = bearerPattern.ReplaceAllString(out, "Bearer "+redactedPlaceholder)
	out = envLinePattern.ReplaceAllString(out, "${1}="+redactedPlaceholder)
	out = skPattern.ReplaceAllString(out, redactedPlaceholder)
	return out
}

// ShouldRedactFile reports whether a path likely contains secrets.
func ShouldRedactFile(name string) bool {
	base := strings.ToLower(name)
	if strings.HasSuffix(base, ".env") || base == ".env" {
		return true
	}
	for _, frag := range []string{"credentials", "id_rsa", "secret", "token"} {
		if strings.Contains(base, frag) {
			return true
		}
	}
	return false
}
