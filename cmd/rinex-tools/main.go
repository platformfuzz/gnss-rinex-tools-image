package main

import (
	"fmt"
	"os"

	"github.com/platformfuzz/gnss-rinex-tools-image/internal/cli"
)

func main() {
	root := cli.NewRoot()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
