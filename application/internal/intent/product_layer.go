package intent

import "fmt"

// ProductLayerNonInteractiveMessage explains how to proceed without a TTY.
func ProductLayerNonInteractiveMessage(instruction string) string {
	return fmt.Sprintf(`Product-level intent detected.

Run with:
  asa work %q --yes

or inspect first:

  asa work %q --dry-run`, instruction, instruction)
}
