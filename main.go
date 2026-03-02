package main

import (
	"os"

	"github.com/nusii/nusii-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
