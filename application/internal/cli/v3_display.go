package cli

import "os"

func mustWd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
