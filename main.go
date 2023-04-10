package main

import (
	"github.com/hybridgroup/go-sayanything/cmd"
)

var version = "dev"

func main() {
	cmd.RunCLI(version)
}
