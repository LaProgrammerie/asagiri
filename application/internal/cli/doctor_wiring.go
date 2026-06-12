package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/doctorcli"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return doctorcli.RootCommand()
}
