package main

import (
	"os"

	"github.com/jaeyeom/gh-repox/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
