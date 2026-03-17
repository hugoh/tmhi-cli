// Package main is the entry point for tmhi-cli.
package main

import (
	"github.com/hugoh/tmhi-cli/internal"
)

var version = "dev"

func main() {
	internal.Cmd(version)
}
