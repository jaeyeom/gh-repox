package main

import (
	"errors"
	"os"

	"github.com/jaeyeom/gh-repox/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		var exitErr *cli.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
