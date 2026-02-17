// Package main is the entry point for the forg CLI.
package main

import (
	"os"

	"github.com/devaloi/forg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
