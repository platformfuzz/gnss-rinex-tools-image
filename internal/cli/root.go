// Package cli wires the rinex-tools multi-command CLI.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/platformfuzz/gnss-rinex-tools-image/internal/check"
	"github.com/spf13/cobra"
)

// NewRoot builds the root command with all tool subcommands.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "rinex-tools",
		Short:         "GNSS RINEX QC and helper tools",
		Long:          "Container-friendly CLI toolbox for GNSS RINEX observation checks and related helpers.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newCheckCmd(os.Stdout, os.Stderr))
	return root
}

func newCheckCmd(stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "check <file.rnx>",
		Short: "Custom quick RINEX 3 OBS sanity check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			r, err := check.AnalyzeFile(path)
			if err != nil {
				fmt.Fprintln(stderr, err.Error())
				os.Exit(check.ExitCode(nil, err))
			}
			check.WriteReport(stdout, r)
			code := check.ExitCode(r, nil)
			if code != 0 {
				os.Exit(code)
			}
			return nil
		},
	}
}
