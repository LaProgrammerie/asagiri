package safeid

import (
	"fmt"
	"strings"
)

// Validate rejects trust IDs that could escape .asagiri/trust/<id>/ via path segments.
func Validate(id string) error {
	if id == "" {
		return fmt.Errorf("trust id required")
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("invalid trust id: path segments not allowed")
	}
	return nil
}
