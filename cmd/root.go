package cmd

import (
	"fmt"
	"os"

	"github.com/michielvha/crossplane-docs/internal/engine"
	"github.com/spf13/cobra"
)

var (
	// Version information (set via ldflags during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// NewRootCommand returns a fresh command tree for each invocation.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "crossplane-docs",
		Short:         "Generate documentation for Crossplane resources",
		Long:          `Generate configurable API and composition documentation from Crossplane manifests.`,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(newXRDCommand(), newCompositionCommand(), newBatchCommand())
	return root
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	root := NewRootCommand()
	command, err := root.ExecuteC()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		code := 1
		if command != nil {
			check, _ := command.Flags().GetBool("check")
			if check && !engine.IsDrift(err) {
				code = 2
			}
		}
		os.Exit(code)
	}
}
