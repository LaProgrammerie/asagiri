package executiongraph

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var graphIDPattern = regexp.MustCompile(`^graph-\d{4}-\d{2}-\d{2}-[a-z0-9][a-z0-9-]{2,11}$`)

// NewGraphID returns a unique graph identifier (spec §7 example shape).
func NewGraphID() string {
	now := time.Now().UTC()
	suffix := uuid.New().String()[:8]
	return fmt.Sprintf("graph-%s-%s", now.Format("2006-01-02"), suffix)
}

// ValidateGraphID rejects IDs that could escape .asagiri/graphs/<id>/ via path segments.
func ValidateGraphID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: id required", ErrInvalidGraphID)
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("%w: path segments not allowed", ErrInvalidGraphID)
	}
	if !graphIDPattern.MatchString(id) {
		return fmt.Errorf("%w: expected graph-YYYY-MM-DD-<suffix>", ErrInvalidGraphID)
	}
	return nil
}
