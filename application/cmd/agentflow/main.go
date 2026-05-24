package main

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
