package main

import (
	"os"

	"github.com/jasonaloi/forg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
