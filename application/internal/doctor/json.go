package doctor

import (
	"encoding/json"
	"fmt"
	"io"
)

// FormatJSON writes the doctor report as JSON.
func FormatJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("doctor json: %w", err)
	}
	return nil
}
