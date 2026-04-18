// Package main is the entry point for tmhi-cli.
package main

import (
	"os"

	"github.com/hugoh/tmhi-cli/internal"
)

var version = "dev"

func main() {
	err := internal.Cmd(version)
	if err != nil {
		os.Exit(1)
	}
}
