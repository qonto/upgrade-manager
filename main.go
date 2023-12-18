package main

import (
	"os"

	"github.com/qonto/upgrade-manager/cmd"
)

func main() {
	if err := cmd.InitAndRunCommand(); err != nil {
		os.Exit(3)
	}
}
