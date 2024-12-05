package main

import (
	"github.com/hugoh/tmhi-cli/internal"
)

var version = "dev"

func main() {
	internal.Cmd(version)
}
