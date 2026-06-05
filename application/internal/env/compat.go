package env

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var warned sync.Map

func warnLegacy(oldKey, newKey string) {
	if _, loaded := warned.LoadOrStore(oldKey, true); loaded {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "warning: %s is deprecated; use %s\n", oldKey, newKey)
}

// Truthy returns true when value is "1" (trimmed).
func Truthy(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) == "1"
}

// DryRunEnabled reads ASA_DRY_RUN, with AGENTFLOW_DRY_RUN fallback.
func DryRunEnabled() bool {
	if Truthy("ASA_DRY_RUN") {
		return true
	}
	if Truthy("AGENTFLOW_DRY_RUN") {
		warnLegacy("AGENTFLOW_DRY_RUN", "ASA_DRY_RUN")
		return true
	}
	return false
}

// YesEnabled reads ASA_YES, with AGENTFLOW_YES fallback.
func YesEnabled() bool {
	if Truthy("ASA_YES") {
		return true
	}
	if Truthy("AGENTFLOW_YES") {
		warnLegacy("AGENTFLOW_YES", "ASA_YES")
		return true
	}
	return false
}
