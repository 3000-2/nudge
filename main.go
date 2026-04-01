package main

import (
	"fmt"
	"os"

	"github.com/3000-2/nudge/cmd"
)

var version = "dev"

func main() {
	if err := cmd.NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
