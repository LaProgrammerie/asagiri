package replaycli

import "os"

func osGetwdMust() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}
