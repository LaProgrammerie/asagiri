package replay

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var replayIDPattern = regexp.MustCompile(`^replay-\d{4}-\d{2}-\d{2}-[a-z0-9][a-z0-9-]{2,11}$`)

// NewReplayID returns a unique replay identifier (spec §7).
func NewReplayID() string {
	now := time.Now().UTC()
	suffix := uuid.New().String()[:8]
	return fmt.Sprintf("replay-%s-%s", now.Format("2006-01-02"), suffix)
}

// ValidateReplayID rejects IDs that could escape .asagiri/replays/<id>/ via path segments.
func ValidateReplayID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: id required", ErrInvalidReplayID)
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("%w: path segments not allowed", ErrInvalidReplayID)
	}
	if !replayIDPattern.MatchString(id) {
		return fmt.Errorf("%w: expected replay-YYYY-MM-DD-<suffix>", ErrInvalidReplayID)
	}
	return nil
}
